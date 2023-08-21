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
	"sync/atomic"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// liburing: sq_ring_needs_enter
func (ring *Ring) sqRingNeedsEnter(submit uint32, flags *uint32) bool {
	if submit == 0 {
		return false
	}

	if (ring.flags & SetupSQPoll) == 0 {
		return true
	}

	if atomic.LoadUint32(ring.sqRing.flags)&SQNeedWakeup != 0 {
		*flags |= EnterSQWakeup

		return true
	}

	return false
}

// liburing: cq_ring_needs_flush
func (ring *Ring) cqRingNeedsFlush() bool {
	return atomic.LoadUint32(ring.sqRing.flags)&(SQCQOverflow|SQTaskrun) != 0
}

// liburing: cq_ring_needs_enter
func (ring *Ring) cqRingNeedsEnter() bool {
	return (ring.flags&SetupIOPoll) != 0 || ring.cqRingNeedsFlush()
}

// liburing: get_data
type getData struct {
	submit   uint32
	waitNr   uint32
	getFlags uint32
	sz       int
	hasTS    bool
	arg      unsafe.Pointer
}

// liburing: _io_uring_get_cqe
func (ring *Ring) privateGetCQE(data *getData) (*CompletionQueueEvent, error) {
	var cqe *CompletionQueueEvent
	var looped bool
	var err error

	for {
		var needEnter bool
		var flags uint32
		var nrAvailable uint32
		var ret uint
		var localErr error

		cqe, localErr = internalPeekCQE(ring, &nrAvailable)
		if localErr != nil {
			if err == nil {
				err = localErr
			}

			break
		}
		if cqe == nil && data.waitNr == 0 && data.submit == 0 {
			if looped || !ring.cqRingNeedsEnter() {
				if err == nil {
					err = unix.EAGAIN
				}

				break
			}
			needEnter = true
		}
		if data.waitNr > nrAvailable || needEnter {
			flags = EnterGetEvents | data.getFlags
			needEnter = true
		}
		if ring.sqRingNeedsEnter(data.submit, &flags) {
			needEnter = true
		}
		if !needEnter {
			break
		}
		if looped && data.hasTS {
			arg := (*GetEventsArg)(data.arg)
			if cqe == nil && arg.ts != 0 && err == nil {
				err = unix.ETIME
			}

			break
		}
		if ring.intFlags&IntFlagRegRing != 0 {
			flags |= EnterRegisteredRing
		}
		ret, localErr = ring.Enter2(data.submit, data.waitNr, flags, data.arg, data.sz)
		if localErr != nil {
			if err == nil {
				err = localErr
			}

			break
		}
		data.submit -= uint32(ret)
		if cqe != nil {
			break
		}
		if !looped {
			looped = true
			err = localErr
		}
	}

	return cqe, err
}

// liburing: __io_uring_get_cqe
func (ring *Ring) internalGetCQE(submit uint32, waitNr uint32, sigmask *unix.Sigset_t) (*CompletionQueueEvent, error) {
	data := getData{
		submit:   submit,
		waitNr:   waitNr,
		getFlags: 0,
		sz:       nSig / szDivider,
		arg:      unsafe.Pointer(sigmask),
	}

	cqe, err := ring.privateGetCQE(&data)
	runtime.KeepAlive(data)

	return cqe, err
}

// liburing: io_uring_get_events - https://manpages.debian.org/unstable/liburing-dev/io_uring_get_events.3.en.html
func (ring *Ring) GetEvents() (uint, error) {
	flags := EnterGetEvents

	if ring.intFlags&IntFlagRegRing != 0 {
		flags |= EnterRegisteredRing
	}

	return ring.Enter(0, 0, flags, nil)
}

// liburing: io_uring_peek_batch_cqe
func (ring *Ring) PeekBatchCQE(cqes []*CompletionQueueEvent) uint32 {
	var ready uint32
	var overflowChecked bool
	var shift int

	if ring.flags&SetupCQE32 != 0 {
		shift = 1
	}

	count := uint32(len(cqes))

again:
	ready = ring.CQReady()
	if ready != 0 {
		head := *ring.cqRing.head
		mask := *ring.cqRing.ringMask
		last := head + count
		if count > ready {
			count = ready
		}
		for i := 0; head != last; head, i = head+1, i+1 {
			cqes[i] = (*CompletionQueueEvent)(
				unsafe.Add(
					unsafe.Pointer(ring.cqRing.cqes),
					uintptr((head&mask)<<shift)*unsafe.Sizeof(CompletionQueueEvent{}),
				),
			)
		}

		return count
	}

	if overflowChecked {
		return 0
	}

	if ring.cqRingNeedsFlush() {
		_, _ = ring.GetEvents()
		overflowChecked = true

		goto again
	}

	return 0
}

// liburing: __io_uring_flush_sq
func (ring *Ring) internalFlushSQ() uint32 {
	sq := ring.sqRing
	tail := sq.sqeTail

	if sq.sqeHead != tail {
		sq.sqeHead = tail
		atomic.StoreUint32(sq.tail, tail)
	}

	return tail - atomic.LoadUint32(sq.head)
}

// liburing: io_uring_wait_cqes_new
func (ring *Ring) WaitCQEsNew(
	waitNr uint32, ts *syscall.Timespec, sigmask *unix.Sigset_t,
) (*CompletionQueueEvent, error) {
	var arg *GetEventsArg
	var data *getData

	arg = &GetEventsArg{
		sigMask:   uint64(uintptr(unsafe.Pointer(sigmask))),
		sigMaskSz: nSig / szDivider,
		ts:        uint64(uintptr(unsafe.Pointer(ts))),
	}

	data = &getData{
		waitNr:   waitNr,
		getFlags: EnterExtArg,
		sz:       int(unsafe.Sizeof(GetEventsArg{})),
		hasTS:    true,
		arg:      unsafe.Pointer(arg),
	}

	cqe, err := ring.privateGetCQE(data)
	runtime.KeepAlive(data)

	return cqe, err
}

// liburing: __io_uring_submit_timeout
func (ring *Ring) internalSubmitTimeout(waitNr uint32, ts *syscall.Timespec) (uint32, error) {
	var sqe *SubmissionQueueEntry
	var err error

	/*
	 * If the SQ ring is full, we may need to submit IO first
	 */
	sqe = ring.GetSQE()
	if sqe == nil {
		_, err = ring.Submit()
		if err != nil {
			return 0, err
		}
		sqe = ring.GetSQE()
		if sqe == nil {
			return 0, syscall.EAGAIN
		}
	}
	sqe.PrepareTimeout(ts, waitNr, 0)
	sqe.UserData = liburingUdataTimeout

	return ring.internalFlushSQ(), nil
}

// liburing: io_uring_wait_cqes - https://manpages.debian.org/unstable/liburing-dev/io_uring_wait_cqes.3.en.html
func (ring *Ring) WaitCQEs(waitNr uint32, ts *syscall.Timespec, sigmask *unix.Sigset_t) (*CompletionQueueEvent, error) {
	var toSubmit uint32
	var err error

	if ts != nil {
		if ring.features&FeatExtArg != 0 {
			return ring.WaitCQEsNew(waitNr, ts, sigmask)
		}
		toSubmit, err = ring.internalSubmitTimeout(waitNr, ts)
		if err != nil {
			return nil, err
		}
	}

	return ring.internalGetCQE(toSubmit, waitNr, sigmask)
}

// liburing: io_uring_submit_and_wait_timeout - https://manpages.debian.org/unstable/liburing-dev/SubmitAndWaitTimeout.3.en.html
func (ring *Ring) SubmitAndWaitTimeout(
	waitNr uint32, ts *syscall.Timespec, sigmask *unix.Sigset_t,
) (*CompletionQueueEvent, error) {
	var toSubmit uint32
	var err error
	var cqe *CompletionQueueEvent

	if ts != nil {
		if ring.features&FeatExtArg != 0 {
			arg := GetEventsArg{
				sigMask:   uint64(uintptr(unsafe.Pointer(sigmask))),
				sigMaskSz: nSig / szDivider,
				ts:        uint64(uintptr(unsafe.Pointer(ts))),
			}
			data := getData{
				submit:   ring.internalFlushSQ(),
				waitNr:   waitNr,
				getFlags: EnterExtArg,
				sz:       int(unsafe.Sizeof(arg)),
				hasTS:    ts != nil,
				arg:      unsafe.Pointer(&arg),
			}

			cqe, err = ring.privateGetCQE(&data)
			runtime.KeepAlive(data)

			return cqe, err
		}
		toSubmit, err = ring.internalSubmitTimeout(waitNr, ts)
		if err != nil {
			return cqe, err
		}
	} else {
		toSubmit = ring.internalFlushSQ()
	}

	return ring.internalGetCQE(toSubmit, waitNr, sigmask)
}

// liburing: io_uring_wait_cqe_timeout - https://manpages.debian.org/unstable/liburing-dev/io_uring_wait_cqe_timeout.3.en.html
func (ring *Ring) WaitCQETimeout(ts *syscall.Timespec) (*CompletionQueueEvent, error) {
	return ring.WaitCQEs(1, ts, nil)
}

// liburing: __io_uring_submit
func (ring *Ring) internalSubmit(submitted uint32, waitNr uint32, getEvents bool) (uint, error) {
	cqNeedsEnter := getEvents || waitNr != 0 || ring.cqRingNeedsEnter()

	var flags uint32
	var ret uint
	var err error

	flags = 0
	if ring.sqRingNeedsEnter(submitted, &flags) || cqNeedsEnter {
		if cqNeedsEnter {
			flags |= EnterGetEvents
		}
		if ring.intFlags&IntFlagRegRing != 0 {
			flags |= EnterRegisteredRing
		}

		ret, err = ring.Enter(submitted, waitNr, flags, nil)
		if err != nil {
			return 0, err
		}
	} else {
		ret = uint(submitted)
	}

	return ret, nil
}

// liburing: __io_uring_submit_and_wait
func (ring *Ring) internalSubmitAndWait(waitNr uint32) (uint, error) {
	return ring.internalSubmit(ring.internalFlushSQ(), waitNr, false)
}

// liburing: io_uring_submit - https://manpages.debian.org/unstable/liburing-dev/io_uring_submit.3.en.html
func (ring *Ring) Submit() (uint, error) {
	return ring.internalSubmitAndWait(0)
}

// liburing: io_uring_submit_and_wait - https://manpages.debian.org/unstable/liburing-dev/io_uring_submit_and_wait.3.en.html
func (ring *Ring) SubmitAndWait(waitNr uint32) (uint, error) {
	return ring.internalSubmitAndWait(waitNr)
}

// liburing: io_uring_submit_and_get_events - https://manpages.debian.org/unstable/liburing-dev/io_uring_submit_and_get_events.3.en.html
func (ring *Ring) SubmitAndGetEvents() (uint, error) {
	return ring.internalSubmit(ring.internalFlushSQ(), 0, true)
}

// __io_uring_sqring_wait
func (ring *Ring) internalSQRingWait() (uint, error) {
	flags := EnterSQWait

	if ring.intFlags&IntFlagRegRegRing != 0 {
		flags |= EnterRegisteredRing
	}

	return ring.Enter(0, 0, flags, nil)
}
