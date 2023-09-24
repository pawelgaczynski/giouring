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

	"golang.org/x/sys/unix"

	. "github.com/stretchr/testify/require"
)

func createSocketPair(t *testing.T, buffer []byte) [2]int {
	socketPair, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	NoError(t, err)

	err = syscall.SetsockoptInt(socketPair[0], syscall.SOL_SOCKET, syscall.SO_SNDBUF, 1)
	NoError(t, err)

	err = unix.Send(socketPair[0], buffer, 0)
	NoError(t, err)

	return socketPair
}

var setupPoll = func(t *testing.T, ring *Ring, ctx testContext) {
	buffer := []byte("xy")
	ctx["buffer"] = buffer

	socketPair := createSocketPair(t, buffer)
	ctx["socketPair"] = socketPair
}

var pollAction = func(actions int) action {
	return func(t *testing.T, ctx testContext) int {
		socketPair, ok := ctx["socketPair"].([2]int)
		True(t, ok)
		buffer, ok := ctx["buffer"].([]byte)
		True(t, ok)
		err := unix.Send(socketPair[1], buffer, 0)
		NoError(t, err)

		_, err = unix.Read(socketPair[1], buffer)
		NoError(t, err)

		return actions
	}
}

var pollCleanup = func(ctx testContext) {
	if val, ok := ctx["socketPair"]; ok {
		if socketPair, socketOk := val.([2]int); socketOk {
			syscall.Close(socketPair[0])
			syscall.Close(socketPair[1])
		}
	}
}

func TestPollMultishot(t *testing.T) {
	testCase(t, testScenario{
		setup: setupPoll,

		prepares: []prepare{
			func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
				socketPair, ok := ctx["socketPair"].([2]int)
				True(t, ok)
				sqe.PreparePollMultishot(socketPair[0], unix.POLLIN|unix.POLLOUT)
				sqe.UserData = 1
			},
		},

		action: pollAction(2),

		result: func(t *testing.T, ctx testContext, cqes []*CompletionQueueEvent) {
			for _, cqe := range cqes {
				Equal(t, uint64(1), cqe.UserData)
				Contains(t, []int32{1, 4}, cqe.Res)
				if (cqe.Res & unix.POLLIN) != 0 {
					ctx["pollinDone"] = true
				} else if (cqe.Res & unix.POLLOUT) != 0 {
					ctx["polloutDone"] = true
				}
			}
		},

		assert: func(t *testing.T, ctx testContext) {
			val, ok := ctx["pollinDone"].(bool)
			True(t, ok)
			True(t, val)
			val, ok = ctx["polloutDone"].(bool)
			True(t, ok)
			True(t, val)
		},

		cleanup: pollCleanup,
	})
}

func TestPollAdd(t *testing.T) {
	testCase(t, testScenario{
		setup: setupPoll,

		prepares: []prepare{
			func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
				socketPair, ok := ctx["socketPair"].([2]int)
				True(t, ok)
				sqe.PreparePollAdd(socketPair[0], unix.POLLIN)
				sqe.UserData = 1
			},
		},

		action: pollAction(1),

		result: func(t *testing.T, ctx testContext, cqes []*CompletionQueueEvent) {
			for _, cqe := range cqes {
				Equal(t, uint64(1), cqe.UserData)
				Equal(t, unix.POLLIN, int(cqe.Res))
			}
		},

		cleanup: pollCleanup,
	})
}

func TestPollUpdate(t *testing.T) {
	ring, err := CreateRing(8)
	NotNil(t, ring)
	NoError(t, err)

	defer ring.QueueExit()

	buffer := []byte("xy")
	socketPair := createSocketPair(t, buffer)

	sqe := ring.GetSQE()
	sqe.PreparePollMultishot(socketPair[0], unix.POLLIN)
	sqe.UserData = 1

	submitted, err := ring.Submit()
	NoError(t, err)
	Equal(t, uint(1), submitted)

	sqe = ring.GetSQE()
	sqe.PreparePollUpdate(1, 3, unix.POLLOUT, PollUpdateEvents|PollUpdateUserData)
	sqe.UserData = 2

	submitted, err = ring.SubmitAndWait(1)
	NoError(t, err)
	Equal(t, uint(1), submitted)

	cqe, err := ring.PeekCQE()
	NoError(t, err)
	NotNil(t, cqe)

	Equal(t, uint64(2), cqe.UserData)
	Equal(t, 0, int(cqe.Res))
	ring.CQESeen(cqe)

	err = unix.Send(socketPair[1], buffer, 0)
	NoError(t, err)

	cqe, err = ring.WaitCQENr(1)
	NoError(t, err)
	NotNil(t, cqe)

	Equal(t, uint64(3), cqe.UserData)
	Equal(t, int32(unix.POLLOUT), cqe.Res)
	Equal(t, CQEFMore, cqe.Flags)
}

func TestPollRemove(t *testing.T) {
	ring, err := CreateRing(8)
	NotNil(t, ring)
	NoError(t, err)

	defer ring.QueueExit()

	buffer := []byte("xy")
	socketPair := createSocketPair(t, buffer)

	sqe := ring.GetSQE()
	sqe.PreparePollAdd(socketPair[0], unix.POLLIN)
	sqe.UserData = 1

	submitted, err := ring.Submit()
	NoError(t, err)
	Equal(t, uint(1), submitted)

	sqe = ring.GetSQE()
	sqe.PreparePollRemove(1)
	sqe.UserData = 2

	submitted, err = ring.SubmitAndWait(1)
	NoError(t, err)
	Equal(t, uint(1), submitted)

	cqe, err := ring.PeekCQE()
	NoError(t, err)
	NotNil(t, cqe)
	Equal(t, int32(syscall.ECANCELED), -cqe.Res)
	Equal(t, uint64(1), cqe.UserData)
	ring.CQESeen(cqe)

	cqe, err = ring.PeekCQE()
	NoError(t, err)
	NotNil(t, cqe)
	Equal(t, int32(0), cqe.Res)
	Equal(t, uint64(2), cqe.UserData)
	ring.CQESeen(cqe)

	err = unix.Send(socketPair[1], buffer, 0)
	NoError(t, err)

	time := &syscall.Timespec{Sec: 0, Nsec: 100000000}
	cqe, err = ring.WaitCQETimeout(time)
	ErrorIs(t, err, syscall.ETIME)
	Nil(t, cqe)
}
