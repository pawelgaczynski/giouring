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
	"runtime"
	"syscall"
	"testing"
	"time"

	. "github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestLinkTimeout(t *testing.T) {
	ring, err := CreateRing(8)
	NotNil(t, ring)
	NoError(t, err)

	defer ring.QueueExit()

	buffer := []byte("xy")
	socketPair := createSocketPair(t, buffer)

	sqe := ring.GetSQE()
	sqe.PreparePollMultishot(socketPair[0], unix.POLLIN)
	sqe.UserData = 1
	sqe.Flags |= SqeIOLink

	timespec := &syscall.Timespec{
		Nsec: 200000000,
	}
	sqe = ring.GetSQE()
	sqe.PrepareLinkTimeout(timespec, 0)
	sqe.UserData = 2

	startTime := time.Now()
	submitted, err := ring.Submit()
	runtime.KeepAlive(timespec)
	NoError(t, err)
	Equal(t, uint(2), submitted)

	Equal(t, uint32(0), ring.CQReady())

	cqe, err := ring.WaitCQE()
	NoError(t, err)
	NotNil(t, cqe)
	Equal(t, uint64(2), cqe.UserData)
	Equal(t, int32(syscall.ETIME), -cqe.Res)
	ring.CQESeen(cqe)

	cqe, err = ring.PeekCQE()
	NoError(t, err)
	NotNil(t, cqe)
	Equal(t, uint64(1), cqe.UserData)
	Equal(t, int32(syscall.ECANCELED), -cqe.Res)
	elapsed := time.Since(startTime)
	Greater(t, elapsed, time.Millisecond*200)
	Less(t, elapsed, time.Millisecond*300)
}

func TestTimeout(t *testing.T) {
	ring, err := CreateRing(8)
	NotNil(t, ring)
	NoError(t, err)

	defer ring.QueueExit()

	timespec := &syscall.Timespec{
		Nsec: 200000000,
	}
	sqe := ring.GetSQE()
	sqe.PrepareTimeout(timespec, 0, 0)
	sqe.UserData = 1

	submitted, err := ring.Submit()
	startTime := time.Now()
	runtime.KeepAlive(timespec)
	NoError(t, err)
	Equal(t, uint(1), submitted)

	Equal(t, uint32(0), ring.CQReady())

	cqe, err := ring.WaitCQE()
	NoError(t, err)
	NotNil(t, cqe)

	Equal(t, uint64(1), cqe.UserData)
	Equal(t, int32(syscall.ETIME), -cqe.Res)
	ring.CQESeen(cqe)

	elapsed := time.Since(startTime)
	Greater(t, elapsed, time.Millisecond*200)
	Less(t, elapsed, time.Millisecond*300)
}

func TestTimeoutUpdate(t *testing.T) {
	ring, err := CreateRing(8)
	NotNil(t, ring)
	NoError(t, err)

	defer ring.QueueExit()

	timespec := &syscall.Timespec{
		Sec: 1,
	}
	sqe := ring.GetSQE()
	sqe.PrepareTimeout(timespec, 0, 0)
	sqe.UserData = 1

	submitted, err := ring.Submit()
	startTime := time.Now()
	runtime.KeepAlive(timespec)
	NoError(t, err)
	Equal(t, uint(1), submitted)

	Equal(t, uint32(0), ring.CQReady())

	newTimespec := &syscall.Timespec{
		Nsec: 200000000,
	}
	sqe = ring.GetSQE()
	sqe.PrepareTimeoutUpdate(newTimespec, 1, 0)
	sqe.UserData = 2

	submitted, err = ring.Submit()
	runtime.KeepAlive(newTimespec)
	NoError(t, err)
	Equal(t, uint(1), submitted)

	Equal(t, uint32(1), ring.CQReady())

	cqe, err := ring.WaitCQENr(1)
	NoError(t, err)
	NotNil(t, cqe)

	Equal(t, uint64(2), cqe.UserData)
	Equal(t, 0, int(cqe.Res))
	ring.CQESeen(cqe)

	cqe, err = ring.WaitCQENr(1)
	NoError(t, err)
	NotNil(t, cqe)

	Equal(t, uint64(1), cqe.UserData)
	Equal(t, int32(syscall.ETIME), -cqe.Res)
	ring.CQESeen(cqe)

	elapsed := time.Since(startTime)
	Greater(t, elapsed, time.Millisecond*200)
	Less(t, elapsed, time.Millisecond*300)
}

func TestTimeoutRemove(t *testing.T) {
	ring, err := CreateRing(8)
	NotNil(t, ring)
	NoError(t, err)

	defer ring.QueueExit()

	timespec := &syscall.Timespec{
		Sec: 1,
	}
	sqe := ring.GetSQE()
	sqe.PrepareTimeout(timespec, 0, 0)
	sqe.UserData = 1

	submitted, err := ring.Submit()
	runtime.KeepAlive(timespec)
	NoError(t, err)
	Equal(t, uint(1), submitted)

	Equal(t, uint32(0), ring.CQReady())

	sqe = ring.GetSQE()
	sqe.PrepareTimeoutRemove(1, 0)
	sqe.UserData = 2

	submitted, err = ring.Submit()
	NoError(t, err)
	Equal(t, uint(1), submitted)

	Equal(t, uint32(2), ring.CQReady())

	cqe, err := ring.WaitCQENr(2)
	NoError(t, err)
	NotNil(t, cqe)

	cqes := make([]*CompletionQueueEvent, 2)
	cqeNr := ring.PeekBatchCQE(cqes)
	Equal(t, uint32(2), cqeNr)

	cqe = cqes[0]
	Equal(t, uint64(2), cqe.UserData)
	Equal(t, 0, int(cqe.Res))
	ring.CQESeen(cqe)

	cqe = cqes[1]
	Equal(t, uint64(1), cqe.UserData)
	Equal(t, int32(syscall.ECANCELED), -cqe.Res)
	ring.CQESeen(cqe)
}
