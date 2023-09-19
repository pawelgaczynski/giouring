// MIT License
//
// Copyright (c) 2023 Paweł Gaczyński
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
// OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
// CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
// TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package giouring

import (
	"syscall"
	"testing"
	"time"
	"unsafe"

	. "github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

var udpTestPort = getTestPort()

func prepareMsgHdr(buffer []byte, rsa *syscall.RawSockaddrAny, addressSize int) syscall.Msghdr {
	var iovec syscall.Iovec
	iovec.Base = (*byte)(unsafe.Pointer(&buffer[0]))
	iovec.SetLen(len(buffer))

	var msg syscall.Msghdr
	msg.Name = (*byte)(unsafe.Pointer(rsa))
	msg.Namelen = uint32(addressSize)
	msg.Iov = &iovec
	msg.Iovlen = 1

	return msg
}

type udpServer struct {
	ring         *Ring
	dataReceived chan bool
}

func (s *udpServer) init(t *testing.T) {
	s.dataReceived = make(chan bool, 1)
	var err error
	s.ring, err = CreateRing(16)
	NoError(t, err)
}

func (s *udpServer) serve(t *testing.T) {
	socketFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	NoError(t, err)
	err = syscall.Bind(socketFd, &syscall.SockaddrInet4{
		Port: udpTestPort,
		Addr: [4]byte{127, 0, 0, 1},
	})
	NoError(t, err)

	defer func() {
		closerErr := syscall.Close(socketFd)
		Nil(t, closerErr)
	}()

	entry := s.ring.GetSQE()
	NotNil(t, entry)

	buffer := make([]byte, 64)
	var rsa syscall.RawSockaddrAny
	msghdr := prepareMsgHdr(buffer, &rsa, syscall.SizeofSockaddrAny)

	entry.PrepareRecvMsg(socketFd, &msghdr, 0)
	entry.UserData = 101

	submitted, err := s.ring.Submit()
	NoError(t, err)
	Equal(t, uint(1), submitted)

	cqe, err := s.ring.WaitCQE()
	NoError(t, err)
	NotNil(t, cqe)

	Equal(t, uint64(101), cqe.UserData)
	Equal(t, cqe.Res, int32(18))

	s.ring.CQESeen(cqe)

	s.dataReceived <- true
}

func (s *udpServer) exit() {
	defer s.ring.QueueExit()
}

type clientCtx struct {
	clientSockFd int
	rsa          *syscall.RawSockaddrAny
	addressSize  int
	buffer       []byte
	msghdr       *syscall.Msghdr
}

func sendtoFunctionFactory(fixed bool) func(*SubmissionQueueEntry, *clientCtx) {
	return func(entry *SubmissionQueueEntry, ctx *clientCtx) {
		entry.PrepareSendto(ctx.clientSockFd, ctx.buffer, 0, ctx.rsa, uint32(ctx.addressSize))
		entry.UserData = 4
		if fixed {
			entry.Flags |= SqeFixedFile
		}
	}
}

func sendmsgFunction(entry *SubmissionQueueEntry, ctx *clientCtx) {
	msghdr := prepareMsgHdr(ctx.buffer, ctx.rsa, ctx.addressSize)
	ctx.msghdr = &msghdr
	entry.PrepareSendMsg(ctx.clientSockFd, ctx.msghdr, 0)
	entry.UserData = 4
}

func prepareSocket(entry *SubmissionQueueEntry) {
	entry.PrepareSocket(unix.AF_INET, unix.SOCK_DGRAM, 0, 0)
}

func prepareSocketDirect(entry *SubmissionQueueEntry) {
	entry.PrepareSocketDirect(unix.AF_INET, unix.SOCK_DGRAM, 0, 0, 0)
}

func prepareSocketDirectAlloc(entry *SubmissionQueueEntry) {
	entry.PrepareSocketDirectAlloc(unix.AF_INET, unix.SOCK_DGRAM, 0, 0)
}

func TestUDPRecvSendto(t *testing.T) {
	testUDP(t, false, prepareSocket, sendtoFunctionFactory(false))
}

func TestUDPRecvSendtoSocketDirect(t *testing.T) {
	testUDP(t, true, prepareSocketDirect, sendtoFunctionFactory(true))
}

func TestUDPRecvSendtoSocketDirectAlloc(t *testing.T) {
	testUDP(t, true, prepareSocketDirectAlloc, sendtoFunctionFactory(true))
}

func TestUDPRecvSendmsg(t *testing.T) {
	testUDP(t, false, prepareSocket, sendmsgFunction)
}

func testUDP(t *testing.T,
	fixed bool,
	prepareSocketFunc func(*SubmissionQueueEntry),
	sendFunc func(*SubmissionQueueEntry, *clientCtx),
) {
	var server udpServer
	server.init(t)

	defer server.exit()

	go func() {
		server.serve(t)
	}()

	ring, err := CreateRing(16)
	NoError(t, err)

	defer ring.QueueExit()

	var files []int
	if fixed {
		files = []int{-1}
		regRes, regErr := ring.RegisterFiles(files)
		NoError(t, regErr)
		Equal(t, uint(0), regRes)
	}

	entry := ring.GetSQE()
	NotNil(t, entry)

	prepareSocketFunc(entry)

	entry.UserData = 1

	submitted, err := ring.SubmitAndWait(1)
	NoError(t, err)
	Equal(t, uint(1), submitted)

	cqe, err := ring.PeekCQE()
	NoError(t, err)
	NotNil(t, cqe)

	Equal(t, uint64(1), cqe.UserData)
	if fixed {
		Equal(t, int32(0), cqe.Res)
	} else {
		Greater(t, cqe.Res, int32(0))
	}

	var ctx clientCtx
	ctx.clientSockFd = int(cqe.Res)

	ring.CQESeen(cqe)

	entry = ring.GetSQE()
	NotNil(t, entry)

	var rawSockAddr syscall.RawSockaddr
	rawSockAddr.Family = syscall.AF_INET
	rawSockAddr.Data = [14]int8{int8(udpTestPort >> 8), int8(udpTestPort), 127, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0}
	ctx.rsa = &syscall.RawSockaddrAny{
		Addr: rawSockAddr,
	}
	ctx.addressSize = syscall.SizeofSockaddrAny

	entry.PrepareConnect(ctx.clientSockFd, ctx.rsa, uint64(ctx.addressSize))
	entry.UserData = 3
	if fixed {
		entry.Flags |= SqeFixedFile
	}

	submitted, err = ring.SubmitAndWait(1)
	NoError(t, err)
	Equal(t, uint(1), submitted)

	cqe, err = ring.PeekCQE()
	NoError(t, err)
	NotNil(t, cqe)

	Equal(t, uint64(3), cqe.UserData)
	Zero(t, cqe.Res)

	ring.CQESeen(cqe)

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	send := func() {
		entry = ring.GetSQE()
		NotNil(t, entry)
		ctx.buffer = []byte("testdata1234567890")

		sendFunc(entry, &ctx)

		submitted, err = ring.SubmitAndWait(1)
		NoError(t, err)
		Equal(t, uint(1), submitted)

		cqe, err = ring.PeekCQE()
		NoError(t, err)
		NotNil(t, cqe)
		Equal(t, uint64(4), cqe.UserData)
		Equal(t, int32(len(ctx.buffer)), cqe.Res)

		ring.CQESeen(cqe)
	}

	send()

	select {
	case <-server.dataReceived:
		return
	default:
		for {
			select {
			case <-server.dataReceived:
				return
			case <-ticker.C:
				send()
			}
		}
	}
}
