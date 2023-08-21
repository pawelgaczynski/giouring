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
	"errors"
	"fmt"
	"net"
	"sync"
	"syscall"
	"testing"
	"time"
	"unsafe"

	. "github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

const (
	acceptFlag uint64 = 1 << (64 - 1 - iota)
	recvFlag
	sendFlag
	closeFlag
)

const allFlagsMask = acceptFlag | recvFlag | sendFlag | closeFlag

const (
	acceptState = iota
	recvState
	sendState
	closeState
)

type tcpConn struct {
	buffer   []byte
	iovecs   []syscall.Iovec
	msg      *syscall.Msghdr
	sockAddr *syscall.RawSockaddrAny
	fd       int
	state    uint8

	receivedCount uint
}

type networkTester struct {
	connections map[int]*tcpConn

	clientAddr *unix.RawSockaddrAny
	clientLen  *uint32

	clientAddrPointer uintptr
	clientLenPointer  uint64

	accpepted uint
}

var prepareSingleAccept = func(t *testing.T, context testContext, entry *SubmissionQueueEntry) {
	socketFd, ok := context["socketFd"].(int)
	True(t, ok)
	clientAddrPointer, ok := context["clientAddrPointer"].(uintptr)
	True(t, ok)
	clientLenPointer, ok := context["clientLenPointer"].(uint64)
	True(t, ok)
	entry.PrepareAccept(socketFd, clientAddrPointer, clientLenPointer, 0)
}

func newNetworkTester() *networkTester {
	clientLen := new(uint32)
	clientAddr := &unix.RawSockaddrAny{}
	*clientLen = unix.SizeofSockaddrAny
	clientAddrPointer := uintptr(unsafe.Pointer(clientAddr))
	clientLenPointer := uint64(uintptr(unsafe.Pointer(clientLen)))

	return &networkTester{
		connections:       make(map[int]*tcpConn),
		clientAddr:        clientAddr,
		clientLen:         clientLen,
		clientAddrPointer: clientAddrPointer,
		clientLenPointer:  clientLenPointer,
	}
}

func (tester *networkTester) getConnection(fd int) *tcpConn {
	if val, ok := tester.connections[fd]; ok {
		return val
	}

	buffer := make([]byte, 1024)
	iovecs := make([]syscall.Iovec, 1)
	iovecs[0] = syscall.Iovec{
		Base: &buffer[0],
		Len:  uint64(len(buffer)),
	}

	var (
		msg      syscall.Msghdr
		sockAddr syscall.RawSockaddrAny
	)

	sizeOfSockAddr := unsafe.Sizeof(sockAddr)
	msg.Name = (*byte)(unsafe.Pointer(&sockAddr))
	msg.Namelen = uint32(sizeOfSockAddr)
	msg.Iov = &iovecs[0]
	msg.Iovlen = 1

	controlLen := cmsgAlign(uint64(sizeOfSockAddr)) + syscall.SizeofCmsghdr
	controlBuffer := make([]byte, controlLen)
	msg.Control = (*byte)(unsafe.Pointer(&controlBuffer[0]))
	msg.SetControllen(int(controlLen))

	connection := &tcpConn{state: acceptState, buffer: buffer, iovecs: iovecs, fd: fd, msg: &msg, sockAddr: &sockAddr}
	tester.connections[fd] = connection

	return connection
}

func (tester *networkTester) loop(
	t *testing.T, scenario networkTestScenario, ctx testContext, ring *Ring, expectedRWLoops uint,
) bool {
	t.Helper()

	socketVal, ok := ctx["socketFd"]
	True(t, ok)
	socketFd, ok := socketVal.(int)
	True(t, ok)

	cqe, err := ring.WaitCQE()
	if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EINTR) ||
		errors.Is(err, syscall.ETIME) {
		return false
	}
	Nil(t, err)

	ring.CQESeen(cqe)

	if cqe.UserData&acceptFlag != 0 {
		tester.accpepted++

		Equal(t, uint64(socketFd), cqe.UserData^acceptFlag)
		GreaterOrEqual(t, cqe.Res, int32(0))
		fd := int(cqe.Res)
		conn := tester.getConnection(fd)

		entry := ring.GetSQE()
		NotNil(t, entry)
		scenario.prepareRecv(t, ctx, conn, entry)

		entry.UserData = recvFlag | uint64(conn.fd)
		conn.state = recvState

		if scenario.repeatAccept && tester.accpepted < scenario.clientsNumber {
			var cqeNr uint
			cqeNr, err = ring.Submit()
			Nil(t, err)
			Equal(t, uint(1), cqeNr)
			entry = ring.GetSQE()
			NotNil(t, entry)
			scenario.prepareAccept(t, ctx, entry)
			entry.UserData = acceptFlag | uint64(socketFd)
		}

		var cqeNr uint
		cqeNr, err = ring.Submit()
		Nil(t, err)
		Equal(t, uint(1), cqeNr)
	} else {
		conn := tester.getConnection(int(cqe.UserData & ^allFlagsMask))

		switch {
		case cqe.UserData&recvFlag != 0:
			conn.receivedCount++

			Equal(t, conn.fd, int(cqe.UserData & ^allFlagsMask))

			var recvLength int32 = 18
			if scenario.recvLengthProvider != nil {
				recvLength = scenario.recvLengthProvider()
			}

			Equal(t, recvLength, cqe.Res)

			dataRecevied := scenario.recvDataProvider(ctx, conn, cqe)
			Equal(t, "testdata1234567890", string(dataRecevied))
			conn.buffer = conn.buffer[:0]
			data := []byte("responsedata0123456789")
			copied := copy(conn.buffer[:len(data)], data)
			Equal(t, 22, copied)
			buffer := conn.buffer[:len(data)]

			entry := ring.GetSQE()
			NotNil(t, entry)
			scenario.prepareSend(t, ctx, conn, buffer, entry)

			entry.UserData = sendFlag | uint64(conn.fd)
			conn.state = sendState
			var cqeNr uint
			cqeNr, err = ring.Submit()
			Nil(t, err)
			Equal(t, uint(1), cqeNr)

		case cqe.UserData&sendFlag != 0:
			Equal(t, uint64(conn.fd), cqe.UserData & ^allFlagsMask)
			Greater(t, cqe.Res, int32(0))

			if conn.receivedCount == expectedRWLoops {
				entry := ring.GetSQE()
				NotNil(t, entry)
				scenario.prepareClose(t, ctx, conn, entry)

				entry.UserData = closeFlag | uint64(conn.fd)
				conn.state = closeState
				var cqeNr uint
				cqeNr, err = ring.Submit()
				Nil(t, err)
				Equal(t, uint(1), cqeNr)
			}
		case cqe.UserData&closeFlag != 0:
			Equal(t, uint64(conn.fd), cqe.UserData & ^allFlagsMask)
			Equal(t, int32(0), cqe.Res)

			delete(tester.connections, conn.fd)

			return true
		}
	}

	return false
}

type networkTestScenario struct {
	ringFlags     uint32
	setup         func(*testing.T, testContext, *Ring)
	clientsNumber uint
	repeatAccept  bool
	rwLoopNumber  uint
	recvMulti     bool

	prepareAccept func(*testing.T, testContext, *SubmissionQueueEntry)
	prepareRecv   func(*testing.T, testContext, *tcpConn, *SubmissionQueueEntry)
	prepareSend   func(*testing.T, testContext, *tcpConn, []byte, *SubmissionQueueEntry)
	prepareClose  func(*testing.T, testContext, *tcpConn, *SubmissionQueueEntry)

	recvDataProvider   func(testContext, *tcpConn, *CompletionQueueEvent) []byte
	recvLengthProvider func() int32
}

func testNetwork(t *testing.T, scenario networkTestScenario) {
	t.Helper()

	var wg sync.WaitGroup
	wg.Add(1)

	var expectedLoopCount uint = 1
	var loopCount uint

	if scenario.clientsNumber > 0 {
		expectedLoopCount = scenario.clientsNumber
	}

	go func() {
		ring := NewRing()
		err := ring.QueueInit(64, scenario.ringFlags)
		NoError(t, err)

		context := make(map[string]interface{})

		if scenario.setup != nil {
			scenario.setup(t, context, ring)
		}

		defer func() {
			ring.QueueExit()
		}()

		socketFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
		Nil(t, err)
		err = syscall.SetsockoptInt(socketFd, syscall.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		Nil(t, err)
		err = syscall.SetsockoptInt(socketFd, syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
		Nil(t, err)
		testPort := getTestPort()

		err = syscall.Bind(socketFd, &syscall.SockaddrInet4{
			Port: testPort,
		})
		Nil(t, err)
		err = syscall.SetNonblock(socketFd, false)
		Nil(t, err)
		err = syscall.Listen(socketFd, 128)
		Nil(t, err)

		defer func() {
			closeErr := syscall.Close(socketFd)
			Nil(t, closeErr)
		}()

		entry := ring.GetSQE()
		NotNil(t, entry)

		tester := newNetworkTester()
		context["clientAddrPointer"] = tester.clientAddrPointer
		context["clientLenPointer"] = tester.clientLenPointer
		context["socketFd"] = socketFd

		scenario.prepareAccept(t, context, entry)
		entry.UserData = acceptFlag | uint64(socketFd)

		cqeNr, err := ring.Submit()
		Nil(t, err)
		Equal(t, uint(1), cqeNr)

		var expectedRWLoops uint = 1
		if scenario.rwLoopNumber > 0 {
			expectedRWLoops = scenario.rwLoopNumber
		}

		go func() {
			for i := 0; i < int(expectedLoopCount); i++ {
				var (
					cErr error
					conn net.Conn
				)
				conn, cErr = net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", testPort), time.Second)
				Nil(t, cErr)
				NotNil(t, conn)

				for j := 0; j < int(expectedRWLoops); j++ {
					var bytesWritten int
					bytesWritten, cErr = conn.Write([]byte("testdata1234567890"))
					Nil(t, cErr)
					Equal(t, 18, bytesWritten)

					var buffer [22]byte
					bytesWritten, cErr = conn.Read(buffer[:])
					Nil(t, cErr)
					Equal(t, 22, bytesWritten)
					Equal(t, "responsedata0123456789", string(buffer[:]))

					if tcpConn, ok := conn.(*net.TCPConn); ok {
						lErr := tcpConn.SetLinger(0)
						Nil(t, lErr)
					}
				}
			}
		}()

		for {
			if tester.loop(t, scenario, context, ring, expectedRWLoops) {
				loopCount++
				if loopCount == expectedLoopCount {
					break
				}
			}
		}
		wg.Done()
	}()

	wg.Wait()
}

func TestAcceptSendRecvTCP(t *testing.T) {
	testNetwork(t, networkTestScenario{
		prepareAccept: prepareSingleAccept,
		prepareRecv: func(t *testing.T, ctx testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareRecv(
				conn.fd, uintptr(unsafe.Pointer(&conn.buffer[0])), uint32(len(conn.buffer)), 0)
		},
		prepareSend: func(t *testing.T, ctx testContext, conn *tcpConn, buffer []byte, entry *SubmissionQueueEntry) {
			entry.PrepareSend(
				conn.fd, uintptr(unsafe.Pointer(&buffer[0])), uint32(len(buffer)), 0)
		},
		prepareClose: func(t *testing.T, ctx testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareClose(conn.fd)
		},

		recvDataProvider: func(ctx testContext, conn *tcpConn, cqe *CompletionQueueEvent) []byte {
			return conn.buffer[:18]
		},
	})
}

func TestAcceptReadWriteTCP(t *testing.T) {
	testNetwork(t, networkTestScenario{
		prepareAccept: prepareSingleAccept,
		prepareRecv: func(t *testing.T, context testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareRead(conn.fd, uintptr(unsafe.Pointer(&conn.buffer[0])), uint32(len(conn.buffer)), 0)
		},
		prepareSend: func(t *testing.T, context testContext, conn *tcpConn, buffer []byte, entry *SubmissionQueueEntry) {
			entry.PrepareWrite(
				conn.fd, uintptr(unsafe.Pointer(&buffer[0])), uint32(len(buffer)), 0)
		},
		prepareClose: func(t *testing.T, tc testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareClose(conn.fd)
		},

		recvDataProvider: func(ctx testContext, conn *tcpConn, cqe *CompletionQueueEvent) []byte {
			return conn.buffer[:18]
		},
	})
}

func TestAcceptReadvWritevTCP(t *testing.T) {
	testNetwork(t, networkTestScenario{
		prepareAccept: prepareSingleAccept,
		prepareRecv: func(t *testing.T, context testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareReadv(conn.fd, uintptr(unsafe.Pointer(&conn.iovecs[0])), uint32(len(conn.iovecs)), 0)
		},
		prepareSend: func(t *testing.T, context testContext, conn *tcpConn, buffer []byte, entry *SubmissionQueueEntry) {
			entry.PrepareWritev(
				conn.fd, uintptr(unsafe.Pointer(&conn.iovecs[0])), uint32(len(conn.iovecs)), 0)
		},
		prepareClose: func(t *testing.T, tc testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareClose(conn.fd)
		},

		recvDataProvider: func(ctx testContext, conn *tcpConn, cqe *CompletionQueueEvent) []byte {
			return conn.buffer[:18]
		},
	})
}

func TestMultiAcceptMultiRecvTCP(t *testing.T) {
	testNetwork(t, networkTestScenario{
		clientsNumber: 2,
		repeatAccept:  false,
		rwLoopNumber:  4,
		setup: func(t *testing.T, ctx testContext, ring *Ring) {
			buffers := make([][]byte, 16)
			ts := syscall.NsecToTimespec((time.Millisecond).Nanoseconds())
			for i := 0; i < len(buffers); i++ {
				buffers[i] = make([]byte, 1024)

				sqe := ring.GetSQE()
				NotNil(t, sqe)

				sqe.PrepareProvideBuffers(uintptr(unsafe.Pointer(&buffers[i][0])), len(buffers[i]), 1, 7, i)
				sqe.UserData = 777

				cqe, err := ring.SubmitAndWaitTimeout(1, &ts, nil)
				NoError(t, err)
				NotNil(t, cqe)

				ring.CQESeen(cqe)
			}
			ctx["buffers"] = buffers
		},
		prepareAccept: func(t *testing.T, context testContext, entry *SubmissionQueueEntry) {
			socketFd, ok := context["socketFd"].(int)
			True(t, ok)
			clientAddrPointer, ok := context["clientAddrPointer"].(uintptr)
			True(t, ok)
			clientLenPointer, ok := context["clientLenPointer"].(uint64)
			True(t, ok)

			entry.PrepareMultishotAccept(socketFd, clientAddrPointer, clientLenPointer, 0)
		},
		prepareRecv: func(t *testing.T, context testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareRecvMultishot(conn.fd, 0, 0, 0)
			entry.Flags |= SqeBufferSelect
			entry.BufIG = 7
		},
		prepareSend: func(t *testing.T, context testContext, conn *tcpConn, buffer []byte, entry *SubmissionQueueEntry) {
			entry.PrepareSend(
				conn.fd, uintptr(unsafe.Pointer(&buffer[0])), uint32(len(buffer)), 0)
		},
		prepareClose: func(t *testing.T, tc testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareClose(conn.fd)
		},

		recvDataProvider: func(ctx testContext, conn *tcpConn, cqe *CompletionQueueEvent) []byte {
			NotZero(t, cqe.Flags&CQEFBuffer)

			bufferIdx := uint16(cqe.Flags >> CQEBufferShift)
			buffers, ok := ctx["buffers"].([][]byte)
			True(t, ok)

			return buffers[bufferIdx][:18]
		},
	})
}

func TestRecvMsgSendMsgTCP(t *testing.T) {
	testNetwork(t, networkTestScenario{
		prepareAccept: prepareSingleAccept,
		prepareRecv: func(t *testing.T, context testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareRecvMsg(conn.fd, conn.msg, 0)
		},
		prepareSend: func(t *testing.T, context testContext, conn *tcpConn, buffer []byte, entry *SubmissionQueueEntry) {
			entry.PrepareSendMsg(conn.fd, conn.msg, 0)
		},
		prepareClose: func(t *testing.T, tc testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareClose(conn.fd)
		},

		recvDataProvider: func(ctx testContext, conn *tcpConn, cqe *CompletionQueueEvent) []byte {
			return conn.buffer[:18]
		},
	})
}

func TestRecvMsgMultiSendTCP(t *testing.T) {
	testNetwork(t, networkTestScenario{
		rwLoopNumber: 4,
		recvMulti:    true,
		setup: func(t *testing.T, ctx testContext, ring *Ring) {
			buffers := make([][]byte, 16)
			ts := syscall.NsecToTimespec((time.Millisecond).Nanoseconds())
			for i := 0; i < len(buffers); i++ {
				buffers[i] = make([]byte, 1024)

				sqe := ring.GetSQE()
				NotNil(t, sqe)

				sqe.PrepareProvideBuffers(uintptr(unsafe.Pointer(&buffers[i][0])), len(buffers[i]), 1, 7, i)
				sqe.UserData = 777

				cqe, err := ring.SubmitAndWaitTimeout(1, &ts, nil)
				NoError(t, err)
				NotNil(t, cqe)

				ring.CQESeen(cqe)
			}
			ctx["buffers"] = buffers
		},
		prepareAccept: prepareSingleAccept,
		prepareRecv: func(t *testing.T, context testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareRecvMsgMultishot(conn.fd, conn.msg, 0)
			entry.Flags |= SqeBufferSelect
			entry.BufIG = 7
		},
		prepareSend: func(t *testing.T, ctx testContext, conn *tcpConn, buffer []byte, entry *SubmissionQueueEntry) {
			entry.PrepareSend(conn.fd, uintptr(unsafe.Pointer(&buffer[0])), uint32(len(buffer)), 0)
		},
		prepareClose: func(t *testing.T, tc testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareClose(conn.fd)
		},

		recvDataProvider: func(ctx testContext, conn *tcpConn, cqe *CompletionQueueEvent) []byte {
			NotZero(t, cqe.Flags&CQEFBuffer)

			bufferIdx := uint16(cqe.Flags >> CQEBufferShift)
			buffers, ok := ctx["buffers"].([][]byte)
			True(t, ok)

			buffer := buffers[bufferIdx]
			recvmsgOut := RecvmsgValidate(unsafe.Pointer(&buffer[0]), int(cqe.Res), conn.msg)
			NotNil(t, recvmsgOut)
			Equal(t, uint32(18), recvmsgOut.PayloadLen)

			// TODO: name validation
			name := recvmsgOut.Name()
			NotNil(t, name)

			payloadLength := recvmsgOut.PayloadLength(int(cqe.Res), conn.msg)
			Equal(t, uint32(18), payloadLength)

			payload := recvmsgOut.Payload(conn.msg)

			return unsafe.Slice((*byte)(payload), payloadLength)
		},
		recvLengthProvider: func() int32 {
			return 274
		},
	})
}

const (
	bufferSize      = 1024
	numberOfBuffers = 16
)

func getBuffer(bufferBase uintptr, idx int) uintptr {
	return bufferBase + uintptr((idx * bufferSize))
}

func recycleBuffer(bufRing *BufAndRing, bufferBase uintptr, idx int) {
	bufRing.BufRingAdd(getBuffer(bufferBase, idx), bufferSize, uint16(idx), BufRingMask(numberOfBuffers), 0)
	bufRing.BufRingAdvance(1)
}

func TestMultiAcceptMultiRecvMultiDirectBufRingTCP(t *testing.T) {
	testNetwork(t, networkTestScenario{
		clientsNumber: 2,
		repeatAccept:  false,
		rwLoopNumber:  4,
		setup: func(t *testing.T, ctx testContext, ring *Ring) {
			fds := make([]int, 16)
			for i := range fds {
				fds[i] = -1
			}
			_, err := ring.RegisterFiles(fds)
			NoError(t, err)
			bufRingSize := int((unsafe.Sizeof(BufAndRing{}) + uintptr(bufferSize)) * uintptr(numberOfBuffers))
			data, err := syscall.Mmap(
				-1,
				0,
				bufRingSize,
				syscall.PROT_READ|syscall.PROT_WRITE,
				syscall.MAP_ANONYMOUS|syscall.MAP_PRIVATE,
			)
			NoError(t, err)
			bufRing := (*BufAndRing)(unsafe.Pointer(&data[0]))

			bufRing.BufRingInit()
			reg := &BufReg{
				RingAddr:    uint64(uintptr(unsafe.Pointer(bufRing))),
				RingEntries: uint32(numberOfBuffers),
				Bgid:        0,
			}
			bufferBase := uintptr(unsafe.Pointer(bufRing)) + uintptr(RingBufStructSize)*uintptr(numberOfBuffers)
			_, err = ring.RegisterBufferRing(reg, 0)
			NoError(t, err)
			for i := 0; i < numberOfBuffers; i++ {
				bufRing.BufRingAdd(getBuffer(bufferBase, i), bufferSize, uint16(i), BufRingMask(uint32(numberOfBuffers)), i)
			}

			bufRing.BufRingAdvance(numberOfBuffers)

			ctx["bufferBase"] = bufferBase
			ctx["bufRing"] = bufRing
			ctx["fds"] = fds
		},
		prepareAccept: func(t *testing.T, context testContext, entry *SubmissionQueueEntry) {
			socketFd, ok := context["socketFd"].(int)
			True(t, ok)

			entry.PrepareMultishotAcceptDirect(socketFd, 0, 0, 0)
		},
		prepareRecv: func(t *testing.T, context testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareRecvMultishot(conn.fd, 0, 0, 0)
			entry.Flags |= SqeBufferSelect
			entry.Flags |= SqeFixedFile
		},
		prepareSend: func(t *testing.T, ctx testContext, conn *tcpConn, buffer []byte, entry *SubmissionQueueEntry) {
			entry.PrepareSend(
				conn.fd, uintptr(unsafe.Pointer(&buffer[0])), uint32(len(buffer)), 0)
			entry.Flags |= SqeFixedFile
		},
		prepareClose: func(t *testing.T, tc testContext, conn *tcpConn, entry *SubmissionQueueEntry) {
			entry.PrepareCloseDirect(uint32(conn.fd))
		},

		recvDataProvider: func(ctx testContext, conn *tcpConn, cqe *CompletionQueueEvent) []byte {
			NotZero(t, cqe.Flags&CQEFBuffer)
			bufferIdx := int(cqe.Flags >> CQEBufferShift)
			bufferBase, ok := ctx["bufferBase"].(uintptr)
			True(t, ok)
			bufRing, ok := ctx["bufRing"].(*BufAndRing)
			True(t, ok)
			bufPtr := getBuffer(bufferBase, bufferIdx)
			ringBuffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), bufferSize)

			buffer := make([]byte, cqe.Res)
			copy(buffer, ringBuffer[:cqe.Res])

			recycleBuffer(bufRing, bufferBase, bufferIdx)

			return buffer
		},
	})
}
