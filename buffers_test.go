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

	. "github.com/stretchr/testify/require"
)

func TestProvideAndRemoveBuffers(t *testing.T) {
	ring, err := CreateRing(16)
	NoError(t, err)
	defer ring.QueueExit()

	buffers := make([][]byte, 16)
	ts := syscall.NsecToTimespec((time.Millisecond).Nanoseconds())
	for i := 0; i < len(buffers); i++ {
		buffers[i] = make([]byte, bufferSize)

		sqe := ring.GetSQE()
		NotNil(t, sqe)

		sqe.PrepareProvideBuffers(buffers[i], len(buffers[i]), 1, i, 0)
		sqe.UserData = 777

		var cqe *CompletionQueueEvent
		cqe, err = ring.SubmitAndWaitTimeout(1, &ts, nil)
		NoError(t, err)
		NotNil(t, cqe)

		Equal(t, int32(0), cqe.Res)
		Equal(t, uint64(777), cqe.UserData)

		ring.CQESeen(cqe)
	}

	for i := 0; i < len(buffers); i++ {
		sqe := ring.GetSQE()
		NotNil(t, sqe)

		sqe.PrepareRemoveBuffers(1, i)
		sqe.UserData = 888

		var cqe *CompletionQueueEvent
		cqe, err = ring.SubmitAndWaitTimeout(1, &ts, nil)
		NoError(t, err)
		NotNil(t, cqe)

		Equal(t, int32(1), cqe.Res)
		Equal(t, uint64(888), cqe.UserData)

		ring.CQESeen(cqe)
	}
}
