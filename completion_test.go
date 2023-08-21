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
	"testing"

	. "github.com/stretchr/testify/require"
)

func queueNOPs(t *testing.T, ring *Ring, number int, offset int) error {
	t.Helper()

	for i := 0; i < number; i++ {
		entry := ring.GetSQE()
		NotNil(t, entry)

		entry.PrepareNop()
		entry.UserData = uint64(i + offset)
	}
	submitted, err := ring.Submit()
	Equal(t, int(submitted), number)

	return err
}

func TestPeekBatchCQE(t *testing.T) {
	ring, err := CreateRing(16)
	NoError(t, err)

	defer ring.QueueExit()

	cqeBuff := make([]*CompletionQueueEvent, 4)

	cnt := ring.PeekBatchCQE(cqeBuff)
	Equal(t, uint32(0), cnt)
	NoError(t, queueNOPs(t, ring, 4, 0))

	cnt = ring.PeekBatchCQE(cqeBuff)
	Equal(t, uint32(4), cnt)

	ring.CQAdvance(4)

	for i := 0; i < 4; i++ {
		Equal(t, uint64(i), cqeBuff[i].UserData)
	}
}

func TestRingCqRingNeedsEnter(t *testing.T) {
	ring := NewRing()
	ring.sqRing = &SubmissionQueue{}

	var flags uint32
	ring.sqRing.flags = &flags

	False(t, ring.cqRingNeedsEnter())

	ring.flags |= SetupIOPoll

	True(t, ring.cqRingNeedsEnter())

	ring.flags = 0

	False(t, ring.cqRingNeedsEnter())

	flags |= SQCQOverflow

	True(t, ring.cqRingNeedsEnter())
}

func TestRingForEachCQE(t *testing.T) {
	entries := 16
	ring, err := CreateRing(uint32(entries))
	NoError(t, err)
	NotNil(t, ring)
	defer ring.QueueExit()

	// Add some events to the submission queue
	for i := 0; i < entries; i++ {
		sqe := ring.GetSQE()
		sqe.PrepareNop()
		sqe.UserData = uint64(i + 1000)
		var submitted uint
		submitted, err = ring.Submit()
		NoError(t, err)
		Equal(t, 1, int(submitted))
	}

	// Wait for the events to complete
	_, err = ring.WaitCQENr(16)
	NoError(t, err)

	// Verify that all events were completed
	count := 0
	ring.ForEachCQE(func(cqe *CompletionQueueEvent) {
		count++
		GreaterOrEqual(t, int(cqe.UserData), 1000)
	})
	Equal(t, entries, count)
}
