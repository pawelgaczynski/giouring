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
)

func TestSubmitAndWait(t *testing.T) {
	ring, err := CreateRing(16)
	NoError(t, err)

	defer ring.QueueExit()

	cqeBuff := make([]*CompletionQueueEvent, 16)

	cnt := ring.PeekBatchCQE(cqeBuff)
	Equal(t, uint32(0), cnt)

	NoError(t, queueNOPs(t, ring, 4, 0))

	timespec := syscall.NsecToTimespec((time.Millisecond * 100).Nanoseconds())
	_, err = ring.SubmitAndWaitTimeout(10, &timespec, nil)
	runtime.KeepAlive(timespec)

	NoError(t, err)
}

func TestSubmitAndWaitNilTimeout(t *testing.T) {
	ring, err := CreateRing(16)
	NoError(t, err)

	defer ring.QueueExit()

	cqeBuff := make([]*CompletionQueueEvent, 16)

	cnt := ring.PeekBatchCQE(cqeBuff)
	Equal(t, uint32(0), cnt)

	NoError(t, queueNOPs(t, ring, 4, 0))

	_, err = ring.SubmitAndWaitTimeout(1, nil, nil)
	NoError(t, err)
}
