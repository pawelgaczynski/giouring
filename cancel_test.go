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

	. "github.com/stretchr/testify/require"
)

func testCancelRequest(t *testing.T, cancelRequest func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) bool) {
	acceptRequest := func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) bool {
		socketFd, ok := ctx["socketFd"].(int)
		True(t, ok)
		sqe.PrepareAccept(socketFd, 0, 0, 0)
		sqe.UserData = 1

		return false
	}

	testCase(t, ringInitParams{}, testScenario{
		setup: func(ctx testContext) {
			socketFd, _ := listenSocket(t)
			ctx["socketFd"] = socketFd
		},

		prepares: []func(*testing.T, testContext, *SubmissionQueueEntry) bool{
			acceptRequest,
			acceptRequest,
			acceptRequest,
			acceptRequest,

			cancelRequest,
		},

		result: func(t *testing.T, ctx testContext, cqes []*CompletionQueueEvent) {
			for _, cqe := range cqes {
				Contains(t, []uint64{1, 2}, cqe.UserData)
				if cqe.UserData == 1 {
					Equal(t, -int32(syscall.ECANCELED), cqe.Res)
				} else {
					Equal(t, int32(4), cqe.Res)
				}
			}
		},

		cleanup: func(ctx testContext) {
			socketFd, ok := ctx["socketFd"].(int)
			True(t, ok)
			syscall.Close(socketFd)
		},
	})
}

func TestCancel64(t *testing.T) {
	cancelRequest := func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) bool {
		socketFd, ok := ctx["socketFd"].(int)
		True(t, ok)
		sqe.PrepareCancel64(0, int(AsyncCancelAll))
		sqe.OpcodeFlags |= AsyncCancelFd
		sqe.Fd = int32(socketFd)
		sqe.UserData = 2

		return true
	}

	testCancelRequest(t, cancelRequest)
}

func TestCancel(t *testing.T) {
	cancelRequest := func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) bool {
		socketFd, ok := ctx["socketFd"].(int)
		True(t, ok)
		sqe.PrepareCancel(0, int(AsyncCancelAll))
		sqe.OpcodeFlags |= AsyncCancelFd
		sqe.Fd = int32(socketFd)
		sqe.UserData = 2

		return true
	}

	testCancelRequest(t, cancelRequest)
}

func TestCancelFd(t *testing.T) {
	cancelRequest := func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) bool {
		socketFd, ok := ctx["socketFd"].(int)
		True(t, ok)
		sqe.PrepareCancelFd(socketFd, AsyncCancelFd|AsyncCancelAll)
		sqe.UserData = 2

		return true
	}

	testCancelRequest(t, cancelRequest)
}
