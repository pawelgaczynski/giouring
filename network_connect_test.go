package giouring

import (
	"fmt"
	"net"
	"strconv"
	"syscall"
	"testing"
	"time"
	"unsafe"

	. "github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestNetworkConnectWrite(t *testing.T) {
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	NoError(t, err)
	_, portStr, err := net.SplitHostPort(listen.Addr().String())
	NoError(t, err)
	port, err := strconv.Atoi(portStr)
	NoError(t, err)
	// t.Logf("running test server at port %d", port)

	data := []byte("testdata1234567890")
	go func() {
		ring, err := CreateRing(16)
		NoError(t, err)
		defer ring.QueueExit()

		// create a TCP socket
		fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
		NoError(t, err)

		// prepare connect
		sqe := ring.GetSQE()
		NotNil(t, sqe)

		sa := syscall.SockaddrInet4{Port: port}
		err = sqe.PrepareConnect(fd, &sa)
		NoError(t, err)
		sqe.UserData = connectFlag

		// submit, wait for connect to complete
		cqe := testSubmitWaitCQE(t, ring, connectFlag)
		True(t, cqe.Res == 0)

		// prepare send
		sqe = ring.GetSQE()
		NotNil(t, sqe)
		sqe.PrepareSend(fd, uintptr(unsafe.Pointer(&data[0])), uint32(len(data)), 0)
		sqe.UserData = sendFlag

		// submit, wait for completion, test completion
		cqe = testSubmitWaitCQE(t, ring, sendFlag)
		True(t, cqe.Res > 0)
		True(t, int(cqe.Res) == len(data))

		// prepare and wait for shutdown
		sqe = ring.GetSQE()
		sqe.PrepareShutdown(fd, 1)
		sqe.UserData = connectFlag
		cqe = testSubmitWaitCQE(t, ring, connectFlag)
		True(t, cqe.Res == 0)
	}()

	// read into readBuffer until eof
	var readBuffer []byte
	conn, err := listen.Accept()
	NoError(t, err)
	chunk := make([]byte, 8)
	for {
		n, err := conn.Read(chunk)
		if n == 0 {
			break
		}
		NoError(t, err)
		True(t, n >= 0)
		readBuffer = append(readBuffer, chunk[:n]...)
	}
	conn.Close()
	listen.Close()

	Equal(t, data, readBuffer)
}

func TestNetworkAcceptRead(t *testing.T) {
	ring, err := CreateRing(16)
	NoError(t, err)
	defer ring.QueueExit()

	// prepare accept
	fd, port := testListen(t)
	// t.Logf("running test server at port %d", port)
	sqe := ring.GetSQE()
	NotNil(t, sqe)
	sqe.PrepareAccept(fd, 0, 0, 0)
	sqe.UserData = acceptFlag

	// tcp connect and write data
	data := []byte("testdata1234567890")
	go func() {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
		Nil(t, err)
		NotNil(t, conn)
		n, err := conn.Write(data)
		NoError(t, err)
		Equal(t, len(data), n)
		conn.Close()
	}()

	// submit and wait for accept to complete, get connection fd
	cqe := testSubmitWaitCQE(t, ring, acceptFlag)
	True(t, cqe.Res > 0)
	connFd := int(cqe.Res)

	// receive in chunks
	var recvBuffer []byte
	chunk := make([]byte, 8)
	for {
		// prepare recv
		sqe := ring.GetSQE()
		NotNil(t, sqe)
		sqe.PrepareRecv(connFd, uintptr(unsafe.Pointer(&chunk[0])), uint32(len(chunk)), 0)
		sqe.UserData = recvFlag

		// submit and wait for recv to complete
		cqe := testSubmitWaitCQE(t, ring, recvFlag)
		True(t, cqe.Res >= 0)

		n := int(cqe.Res)
		// t.Logf("read %d bytes %s", n, chunk[0:n])
		True(t, n >= 0)
		if n == 0 {
			break
		}
		recvBuffer = append(recvBuffer, chunk[:n]...)
	}

	Equal(t, len(data), len(recvBuffer))
	Equal(t, data, recvBuffer)
}

// test helper to submit single sqe wait for completion
// compare completion userdata with expected
func testSubmitWaitCQE(t *testing.T, ring *Ring, userdata uint64) *CompletionQueueEvent {
	waitNr, err := ring.SubmitAndWait(1)
	NoError(t, err)
	Equal(t, uint(1), waitNr)
	cqe, err := ring.PeekCQE()
	NoError(t, err)
	NotNil(t, cqe)
	if cqe.Res < 0 {
		t.Logf("cqe %d %d %d %s", cqe.Res, cqe.UserData, cqe.Flags, syscall.Errno(-cqe.Res))
	}
	True(t, cqe.Res >= 0, "result <0 means error")
	Equal(t, userdata, cqe.UserData, "userdata should match expected operation")
	ring.CQAdvance(1)
	return cqe
}

// create socket and listen on OS assigned port
// return created file descriptor and port
func testListen(t *testing.T) (int, int) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	Nil(t, err)
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	Nil(t, err)
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	Nil(t, err)

	sa := syscall.SockaddrInet4{}
	err = syscall.Bind(fd, &sa)
	Nil(t, err)
	err = syscall.SetNonblock(fd, false)
	Nil(t, err)
	err = syscall.Listen(fd, 128)
	Nil(t, err)

	saAny, err := syscall.Getsockname(fd)
	Nil(t, err)
	port := saAny.(*syscall.SockaddrInet4).Port
	return fd, port
}
