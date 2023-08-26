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
	"fmt"
	"sync"
	"syscall"
	"testing"
	"unsafe"

	. "github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

// const (
// 	udpRecv = iota
// 	udpSend
// )

// func anyToSockaddrInet4(rsa *syscall.RawSockaddrAny) (*syscall.SockaddrInet4, error) {
// 	if rsa == nil {
// 		return nil, syscall.EINVAL
// 	}

// 	if rsa.Addr.Family != syscall.AF_INET {
// 		return nil, syscall.EAFNOSUPPORT
// 	}

// 	rsaPointer := (*syscall.RawSockaddrInet4)(unsafe.Pointer(rsa))
// 	sockAddr := new(syscall.SockaddrInet4)
// 	p := (*[2]byte)(unsafe.Pointer(&rsaPointer.Port))
// 	sockAddr.Port = int(p[0])<<8 + int(p[1])

// 	for i := 0; i < len(sockAddr.Addr); i++ {
// 		sockAddr.Addr[i] = rsaPointer.Addr[i]
// 	}

// 	return sockAddr, nil
// }

// type udpConnection struct {
// 	msg           *syscall.Msghdr
// 	rsa           *syscall.RawSockaddrAny
// 	buffer        []byte
// 	controlBuffer []byte
// 	fd            uint64
// 	state         int
// }

// func udpLoop(t *testing.T, ring *Ring, socketFd int, connection *udpConnection) bool {
// 	t.Helper()

// 	cqe, err := ring.WaitCQE()
// 	if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EINTR) ||
// 		errors.Is(err, syscall.ETIME) {
// 		return false
// 	}

// 	NoError(t, err)
// 	entry := ring.GetSQE()
// 	NotNil(t, entry)
// 	ring.CQESeen(cqe)

// 	switch connection.state {
// 	case udpRecv:
// 		_, err = anyToSockaddrInet4(connection.rsa)
// 		if err != nil {
// 			log.Panic(err)
// 		}

// 		Equal(t, "testdata1234567890", string(connection.buffer[:18]))
// 		connection.buffer = connection.buffer[:0]
// 		data := []byte("responsedata0123456789")
// 		copied := copy(connection.buffer[:len(data)], data)
// 		Equal(t, 22, copied)
// 		buffer := connection.buffer[:len(data)]

// 		connection.msg.Iov.Base = (*byte)(unsafe.Pointer(&buffer[0]))
// 		connection.msg.Iov.SetLen(len(buffer))
// 		entry.PrepareSendMsg(socketFd, connection.msg, 0)

// 		entry.UserData = connection.fd
// 		connection.state = udpSend

// 	case udpSend:
// 		Equal(t, connection.fd, cqe.UserData)
// 		Equal(t, cqe.Res, int32(22))

// 		return true
// 	}
// 	cqeNr, err := ring.Submit()
// 	NoError(t, err)
// 	Equal(t, uint(1), cqeNr)

// 	return false
// }

var udpTestPort = getTestPort()

func prepareMsgHdr(buffer []byte, rsa *syscall.RawSockaddrAny, addressSize int) syscall.Msghdr {
	var iovec syscall.Iovec
	iovec.Base = (*byte)(unsafe.Pointer(&buffer[0]))
	iovec.SetLen(len(buffer))

	var msg syscall.Msghdr
	msg.Name = (*byte)(unsafe.Pointer(rsa))
	msg.Namelen = uint32(addressSize) //uint32(syscall.SizeofSockaddrAny)
	msg.Iov = &iovec
	msg.Iovlen = 1

	// controlBuffer := make([]byte, 1000)
	// msg.Control = (*byte)(unsafe.Pointer(&controlBuffer[0]))
	// msg.SetControllen(len(controlBuffer))
	return msg
}

type udpServer struct {
	ring *Ring
}

func (s *udpServer) init(t *testing.T) {
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

func sendtoFunction(entry *SubmissionQueueEntry, ctx *clientCtx) {
	entry.PrepareSendto(ctx.clientSockFd, ctx.buffer, 0, ctx.rsa, uint32(ctx.addressSize))
	entry.UserData = 4
}

func sendmsgFunction(entry *SubmissionQueueEntry, ctx *clientCtx) {
	msghdr := prepareMsgHdr(ctx.buffer, ctx.rsa, ctx.addressSize)
	ctx.msghdr = &msghdr

	fmt.Printf("msghdr: %+v\n", msghdr)
	fmt.Printf("rsa: %+v\n", ctx.rsa)

	entry.PrepareSendMsg(ctx.clientSockFd, ctx.msghdr, 0)
	entry.UserData = 4
}

func TestUDPRecvSendto(t *testing.T) {
	testUDP(t, sendtoFunction)
}

func TestUDPRecvSendmsg(t *testing.T) {
	testUDP(t, sendmsgFunction)
}

func testUDP(t *testing.T, sendFunc func(*SubmissionQueueEntry, *clientCtx)) {
	var wg sync.WaitGroup
	wg.Add(1)

	var server udpServer
	server.init(t)

	defer server.exit()

	go func() {
		server.serve(t)
		wg.Done()
	}()

	ring, err := CreateRing(16)
	NoError(t, err)

	defer ring.QueueExit()

	entry := ring.GetSQE()
	NotNil(t, entry)

	entry.PrepareSocket(unix.AF_INET, unix.SOCK_DGRAM, 0, 0)
	entry.UserData = 1

	submitted, err := ring.SubmitAndWait(1)
	NoError(t, err)
	Equal(t, uint(1), submitted)

	cqe, err := ring.PeekCQE()
	NoError(t, err)
	NotNil(t, cqe)

	Equal(t, uint64(1), cqe.UserData)
	Greater(t, cqe.Res, int32(0))

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

	// var sockaddr syscall.SockaddrInet4
	// sockaddr.Addr = [4]byte{127, 0, 0, 1}
	// sockaddr.Port = udpTestPort

	entry.PrepareConnect(ctx.clientSockFd, ctx.rsa, uint64(ctx.addressSize))
	// entry.PrepareConnect(clientSockFd, &sockaddr)
	entry.UserData = 3

	submitted, err = ring.SubmitAndWait(1)
	NoError(t, err)
	Equal(t, uint(1), submitted)

	cqe, err = ring.PeekCQE()
	NoError(t, err)
	NotNil(t, cqe)

	Equal(t, uint64(3), cqe.UserData)
	Zero(t, cqe.Res)

	ring.CQESeen(cqe)

	entry = ring.GetSQE()
	NotNil(t, entry)

	ctx.buffer = []byte("testdata1234567890")
	sendFunc(entry, &ctx)

	// sendtoFunction(entry, clientSockFd, rsa, sendBuffer)

	submitted, err = ring.SubmitAndWait(1)
	NoError(t, err)
	Equal(t, uint(1), submitted)

	cqe, err = ring.PeekCQE()
	NoError(t, err)
	NotNil(t, cqe)

	// runtime.KeepAlive(rsa)

	Equal(t, uint64(4), cqe.UserData)
	Equal(t, int32(len(ctx.buffer)), cqe.Res)

	ring.CQESeen(cqe)

	wg.Wait()
}
