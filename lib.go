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
	"sync/atomic"
	"syscall"
	"unsafe"
)

const (
	sysSetup    = 425
	sysEnter    = 426
	sysRegister = 427
)

// liburing: io_uring_sq
type SubmissionQueue struct {
	head        *uint32
	tail        *uint32
	ringMask    *uint32
	ringEntries *uint32
	flags       *uint32
	dropped     *uint32
	array       *uint32
	sqes        *SubmissionQueueEntry

	ringSize uint
	ringPtr  unsafe.Pointer

	sqeHead uint32
	sqeTail uint32

	// nolint: unused
	pad [2]uint32
}

// liburing: io_uring_cq
type CompletionQueue struct {
	head        *uint32
	tail        *uint32
	ringMask    *uint32
	ringEntries *uint32
	flags       *uint32
	overflow    *uint32
	cqes        *CompletionQueueEvent

	ringSize uint
	ringPtr  unsafe.Pointer

	// nolint: unused
	pad [2]uint32
}

// liburing: io_uring
type Ring struct {
	sqRing      *SubmissionQueue
	cqRing      *CompletionQueue
	flags       uint32
	ringFd      int
	features    uint32
	enterRingFd int
	intFlags    uint8
	// nolint: unused
	pad [3]uint8
	// nolint: unused
	pad2 uint32
}

// liburing: io_uring_cqe_shift
func (ring *Ring) cqeShift() uint32 {
	if ring.flags&SetupCQE32 != 0 {
		return 1
	}

	return 0
}

// liburing: io_uring_cqe_index
func (ring *Ring) cqeIndex(ptr, mask uint32) uintptr {
	return uintptr((ptr & mask) << ring.cqeShift())
}

// liburing: io_uring_for_each_cqe - https://manpages.debian.org/unstable/liburing-dev/io_uring_for_each_cqe.3.en.html
func (ring *Ring) ForEachCQE(callback func(cqe *CompletionQueueEvent)) {
	var cqe *CompletionQueueEvent
	for head := atomic.LoadUint32(ring.cqRing.head); ; head++ {
		if head != atomic.LoadUint32(ring.cqRing.tail) {
			cqeIndex := ring.cqeIndex(head, *ring.cqRing.ringMask)
			cqe = (*CompletionQueueEvent)(
				unsafe.Add(unsafe.Pointer(ring.cqRing.cqes), cqeIndex*unsafe.Sizeof(CompletionQueueEvent{})),
			)
			callback(cqe)
		} else {
			break
		}
	}
}

// liburing: io_uring_cq_advance - https://manpages.debian.org/unstable/liburing-dev/io_uring_cq_advance.3.en.html
func (ring *Ring) CQAdvance(numberOfCQEs uint32) {
	atomic.StoreUint32(ring.cqRing.head, *ring.cqRing.head+numberOfCQEs)
}

// liburing: io_uring_cqe_seen - https://manpages.debian.org/unstable/liburing-dev/io_uring_cqe_seen.3.en.html
func (ring *Ring) CQESeen(event *CompletionQueueEvent) {
	if event != nil {
		ring.CQAdvance(1)
	}
}

// liburing: io_uring_sqe_set_data - https://manpages.debian.org/unstable/liburing-dev/io_uring_sqe_set_data.3.en.html
func (entry *SubmissionQueueEntry) SetData(data unsafe.Pointer) {
	entry.UserData = uint64(uintptr(data))
}

// liburing: io_uring_cqe_get_data - https://manpages.debian.org/unstable/liburing-dev/io_uring_cqe_get_data.3.en.html
func (c *CompletionQueueEvent) GetData() unsafe.Pointer {
	return unsafe.Pointer(uintptr(c.UserData))
}

// liburing: io_uring_sqe_set_data64 - https://manpages.debian.org/unstable/liburing-dev/io_uring_sqe_set_data64.3.en.html
func (entry *SubmissionQueueEntry) SetData64(data uint64) {
	entry.UserData = data
}

// liburing: io_uring_cqe_get_data64 - https://manpages.debian.org/unstable/liburing-dev/io_uring_cqe_get_data64.3.en.html
func (c *CompletionQueueEvent) GetData64() uint64 {
	return c.UserData
}

// liburing: io_uring_sqe_set_flags - https://manpages.debian.org/unstable/liburing-dev/io_uring_sqe_set_flags.3.en.html
func (entry *SubmissionQueueEntry) SetFlags(flags uint32) {
	entry.Flags = uint8(flags)
}

// liburing: io_uring_sq_ready - https://manpages.debian.org/unstable/liburing-dev/io_uring_sq_ready.3.en.html
func (ring *Ring) SQReady() uint32 {
	khead := *ring.sqRing.head

	if ring.flags&SetupSQPoll != 0 {
		khead = atomic.LoadUint32(ring.sqRing.head)
	}

	return ring.sqRing.sqeTail - khead
}

// liburing: io_uring_sq_space_left - https://manpages.debian.org/unstable/liburing-dev/io_uring_sq_space_left.3.en.html
func (ring *Ring) SQSpaceLeft() uint32 {
	return *ring.sqRing.ringEntries - ring.SQReady()
}

// liburing: io_uring_sqring_wait - https://manpages.debian.org/unstable/liburing-dev/io_uring_sqring_wait.3.en.html
func (ring *Ring) SQRingWait() (uint, error) {
	if ring.flags&SetupSQPoll == 0 {
		return 0, nil
	}
	if ring.SQSpaceLeft() != 0 {
		return 0, nil
	}

	return ring.internalSQRingWait()
}

// liburing: io_uring_cq_ready - https://manpages.debian.org/unstable/liburing-dev/io_uring_cq_ready.3.en.html
func (ring *Ring) CQReady() uint32 {
	return atomic.LoadUint32(ring.cqRing.tail) - *ring.cqRing.head
}

// liburing: io_uring_cq_has_overflow - https://manpages.debian.org/unstable/liburing-dev/io_uring_cq_has_overflow.3.en.html
func (ring *Ring) CQHasOverflow() bool {
	return atomic.LoadUint32(ring.sqRing.flags)&SQCQOverflow != 0
}

// liburing: io_uring_cq_eventfd_enabled
func (ring *Ring) CQEventfdEnabled() bool {
	if *ring.cqRing.flags == 0 {
		return true
	}

	return !(*ring.cqRing.flags&CQEventFdDisabled != 0)
}

// liburing: io_uring_cq_eventfd_toggle
func (ring *Ring) CqEventfdToggle(enabled bool) error {
	var flags uint32

	if enabled == ring.CQEventfdEnabled() {
		return nil
	}

	if *ring.cqRing.flags == 0 {
		return syscall.EOPNOTSUPP
	}

	flags = *ring.cqRing.flags

	if enabled {
		flags &= ^CQEventFdDisabled
	} else {
		flags |= CQEventFdDisabled
	}

	atomic.StoreUint32(ring.cqRing.flags, flags)

	return nil
}

// liburing: io_uring_wait_cqe_nr - https://manpages.debian.org/unstable/liburing-dev/io_uring_wait_cqe_nr.3.en.html
func (ring *Ring) WaitCQENr(waitNr uint32) (*CompletionQueueEvent, error) {
	return ring.internalGetCQE(0, waitNr, nil)
}

// liburing: __io_uring_peek_cqe
func internalPeekCQE(ring *Ring, nrAvailable *uint32) (*CompletionQueueEvent, error) {
	var cqe *CompletionQueueEvent
	var err error
	var available uint32
	var shift uint32
	mask := *ring.cqRing.ringMask

	if ring.flags&SetupCQE32 != 0 {
		shift = 1
	}

	for {
		tail := atomic.LoadUint32(ring.cqRing.tail)
		head := *ring.cqRing.head

		cqe = nil
		available = tail - head
		if available == 0 {
			break
		}

		cqe = (*CompletionQueueEvent)(
			unsafe.Add(unsafe.Pointer(ring.cqRing.cqes), uintptr((head&mask)<<shift)*unsafe.Sizeof(CompletionQueueEvent{})),
		)

		if ring.features&FeatExtArg == 0 && cqe.UserData == liburingUdataTimeout {
			if cqe.Res < 0 {
				err = syscall.Errno(uintptr(-cqe.Res))
			}
			ring.CQAdvance(1)
			if err == nil {
				continue
			}
			cqe = nil
		}

		break
	}

	if nrAvailable != nil {
		*nrAvailable = available
	}

	return cqe, err
}

// liburing: io_uring_peek_cqe - https://manpages.debian.org/unstable/liburing-dev/io_uring_peek_cqe.3.en.html
func (ring *Ring) PeekCQE() (*CompletionQueueEvent, error) {
	fmt.Println("PeekCQE")
	cqe, err := internalPeekCQE(ring, nil)
	fmt.Printf("internalPeekCQE: %v, %v\n", cqe, err)
	if err == nil && cqe != nil {
		return cqe, nil
	}

	return ring.WaitCQENr(0)
}

// liburing: io_uring_wait_cqe - https://manpages.debian.org/unstable/liburing-dev/io_uring_wait_cqe.3.en.html
func (ring *Ring) WaitCQE() (*CompletionQueueEvent, error) {
	// return ring.WaitCQENr(1)
	cqe, err := internalPeekCQE(ring, nil)
	if err == nil && cqe != nil {
		return cqe, nil
	}

	return ring.WaitCQENr(1)
}

// liburing: _io_uring_get_sqe
func privateGetSQE(ring *Ring) *SubmissionQueueEntry {
	sq := ring.sqRing
	var head, next uint32
	var shift int

	if ring.flags&SetupSQE128 != 0 {
		shift = 1
	}
	head = atomic.LoadUint32(sq.head)
	next = sq.sqeTail + 1
	if next-head <= *sq.ringEntries {
		sqe := (*SubmissionQueueEntry)(
			unsafe.Add(unsafe.Pointer(ring.sqRing.sqes),
				uintptr((sq.sqeTail&*sq.ringMask)<<shift)*unsafe.Sizeof(SubmissionQueueEntry{})),
		)
		sq.sqeTail = next

		return sqe
	}

	return nil
}

// liburing: io_uring_get_sqe - https://manpages.debian.org/unstable/liburing-dev/io_uring_get_sqe.3.en.html
func (ring *Ring) GetSQE() *SubmissionQueueEntry {
	return privateGetSQE(ring)
}
