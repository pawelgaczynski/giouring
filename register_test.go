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

func TestRegisterRingFd(t *testing.T) {
	var maxSQEs uint32 = 1024
	ring, err := CreateRing(maxSQEs)
	Nil(t, err)

	_, err = ring.RegisterRingFd()
	Nil(t, err)

	defer ring.QueueExit()

	var (
		entry *SubmissionQueueEntry
		cqe   *CompletionQueueEvent
	)

	for i := 0; i < 1000; i++ {
		entry = ring.GetSQE()
		NotNil(t, entry)

		entry.PrepareNop()
		entry.UserData = uint64(i)

		timespec := syscall.NsecToTimespec((time.Millisecond).Nanoseconds())
		cqe, err = ring.SubmitAndWaitTimeout(1, &timespec, nil)
		Nil(t, err)
		NotNil(t, cqe)

		runtime.KeepAlive(timespec)

		ring.CQESeen(cqe)
	}
}
