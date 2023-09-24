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
	"math/bits"
	"os"
	"syscall"
	"unsafe"
)

const (
	kernMaxEntries   = 32768
	kernMaxCQEntries = 2 * kernMaxEntries
)

func fls(x int) int {
	if x == 0 {
		return 0
	}

	return 8*int(unsafe.Sizeof(x)) - bits.LeadingZeros32(uint32(x))
}

func roundupPow2(depth uint32) uint32 {
	return 1 << uint32(fls(int(depth-1)))
}

const cqEntriesMultiplier = 2

// liburing: get_sq_cq_entries
func getSqCqEntries(entries uint32, p *Params, sq, cq *uint32) error {
	var cqEntries uint32

	if entries == 0 {
		return syscall.EINVAL
	}
	if entries > kernMaxEntries {
		if p.flags&SetupClamp == 0 {
			return syscall.EINVAL
		}
		entries = kernMaxEntries
	}

	entries = roundupPow2(entries)
	if p.flags&SetupCQSize != 0 {
		if p.cqEntries == 0 {
			return syscall.EINVAL
		}
		cqEntries = p.cqEntries
		if cqEntries > kernMaxCQEntries {
			if p.flags&SetupClamp == 0 {
				return syscall.EINVAL
			}
			cqEntries = kernMaxCQEntries
		}
		cqEntries = roundupPow2(cqEntries)
		if cqEntries < entries {
			return syscall.EINVAL
		}
	} else {
		cqEntries = cqEntriesMultiplier * entries
	}
	*sq = entries
	*cq = cqEntries

	return nil
}

// liburing: io_uring_unmap_rings
func UnmapRings(sq *SubmissionQueue, cq *CompletionQueue) {
	if sq.ringSize > 0 {
		_ = sysMunmap(uintptr(sq.ringPtr), uintptr(sq.ringSize))
	}

	if uintptr(cq.ringPtr) != 0 && cq.ringSize > 0 && cq.ringPtr != sq.ringPtr {
		_ = sysMunmap(uintptr(cq.ringPtr), uintptr(cq.ringSize))
	}
}

// liburing: io_uring_setup_ring_pointers
func SetupRingPointers(p *Params, sq *SubmissionQueue, cq *CompletionQueue) {
	sq.head = (*uint32)(unsafe.Pointer(uintptr(sq.ringPtr) + uintptr(p.sqOff.head)))
	sq.tail = (*uint32)(unsafe.Pointer(uintptr(sq.ringPtr) + uintptr(p.sqOff.tail)))
	sq.ringMask = (*uint32)(unsafe.Pointer(uintptr(sq.ringPtr) + uintptr(p.sqOff.ringMask)))
	sq.ringEntries = (*uint32)(unsafe.Pointer(uintptr(sq.ringPtr) + uintptr(p.sqOff.ringEntries)))
	sq.flags = (*uint32)(unsafe.Pointer(uintptr(sq.ringPtr) + uintptr(p.sqOff.flags)))
	sq.dropped = (*uint32)(unsafe.Pointer(uintptr(sq.ringPtr) + uintptr(p.sqOff.dropped)))
	sq.array = (*uint32)(unsafe.Pointer(uintptr(sq.ringPtr) + uintptr(p.sqOff.array)))

	cq.head = (*uint32)(unsafe.Pointer(uintptr(cq.ringPtr) + uintptr(p.cqOff.head)))
	cq.tail = (*uint32)(unsafe.Pointer(uintptr(cq.ringPtr) + uintptr(p.cqOff.tail)))
	cq.ringMask = (*uint32)(unsafe.Pointer(uintptr(cq.ringPtr) + uintptr(p.cqOff.ringMask)))
	cq.ringEntries = (*uint32)(unsafe.Pointer(uintptr(cq.ringPtr) + uintptr(p.cqOff.ringEntries)))
	cq.overflow = (*uint32)(unsafe.Pointer(uintptr(cq.ringPtr) + uintptr(p.cqOff.overflow)))
	cq.cqes = (*CompletionQueueEvent)(unsafe.Pointer(uintptr(cq.ringPtr) + uintptr(p.cqOff.cqes)))
	if p.cqOff.flags != 0 {
		cq.flags = (*uint32)(unsafe.Pointer(uintptr(cq.ringPtr) + uintptr(p.cqOff.flags)))
	}
}

// liburing: io_uring_mmap
func Mmap(fd int, p *Params, sq *SubmissionQueue, cq *CompletionQueue) error {
	var size uintptr
	var err error

	size = unsafe.Sizeof(CompletionQueueEvent{})
	if p.flags&SetupCQE32 != 0 {
		size += unsafe.Sizeof(CompletionQueueEvent{})
	}

	sq.ringSize = uint(uintptr(p.sqOff.array) + uintptr(p.sqEntries)*unsafe.Sizeof(uint32(0)))
	cq.ringSize = uint(uintptr(p.cqOff.cqes) + uintptr(p.cqEntries)*size)

	if p.features&FeatSingleMMap != 0 {
		if cq.ringSize > sq.ringSize {
			sq.ringSize = cq.ringSize
		}
		cq.ringSize = sq.ringSize
	}

	var ringPtr unsafe.Pointer
	ringPtr, err = sysMmap(0, uintptr(sq.ringSize), syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED|syscall.MAP_POPULATE, fd,
		int64(offsqRing))
	if err != nil {
		return err
	}
	sq.ringPtr = ringPtr

	if p.features&FeatSingleMMap != 0 {
		cq.ringPtr = sq.ringPtr
	} else {
		ringPtr, err = sysMmap(0, uintptr(cq.ringSize), syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_SHARED|syscall.MAP_POPULATE, fd,
			int64(offcqRing))
		if err != nil {
			cq.ringPtr = nil

			goto err
		}
		cq.ringPtr = ringPtr
	}

	size = unsafe.Sizeof(SubmissionQueueEntry{})
	if p.flags&SetupSQE128 != 0 {
		size += 64
	}
	ringPtr, err = sysMmap(0, size*uintptr(p.sqEntries), syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED|syscall.MAP_POPULATE, fd, int64(offSQEs))
	if err != nil {
		goto err
	}
	sq.sqes = (*SubmissionQueueEntry)(ringPtr)
	SetupRingPointers(p, sq, cq)

	return nil

err:
	UnmapRings(sq, cq)

	return err
}

// liburing: io_uring_queue_mmap
func (ring *Ring) QueueMmap(fd int, p *Params) error {
	return Mmap(fd, p, ring.sqRing, ring.cqRing)
}

// liburing: io_uring_ring_dontfork
func (ring *Ring) RingDontFork() error {
	var length uintptr
	var err error

	if ring.sqRing.ringPtr == nil || ring.sqRing.sqes == nil || ring.cqRing.ringPtr == nil {
		return syscall.EINVAL
	}

	length = unsafe.Sizeof(SubmissionQueueEntry{})
	if ring.flags&SetupSQE128 != 0 {
		length += 64
	}
	length *= uintptr(*ring.sqRing.ringEntries)
	err = sysMadvise(uintptr(unsafe.Pointer(ring.sqRing.sqes)), length, syscall.MADV_DONTFORK)
	if err != nil {
		return err
	}

	length = uintptr(ring.sqRing.ringSize)
	err = sysMadvise(uintptr(ring.sqRing.ringPtr), length, syscall.MADV_DONTFORK)
	if err != nil {
		return err
	}

	if ring.cqRing.ringPtr != ring.sqRing.ringPtr {
		length = uintptr(ring.cqRing.ringSize)
		err = sysMadvise(uintptr(ring.cqRing.ringPtr), length, syscall.MADV_DONTFORK)
		if err != nil {
			return err
		}
	}

	return nil
}

/* FIXME */
const hugePageSize uint64 = 2 * 1024 * 1024

// liburing: io_uring_alloc_huge
func allocHuge(
	entries uint32, p *Params, sq *SubmissionQueue, cq *CompletionQueue, buf unsafe.Pointer, bufSize uint64,
) (uint, error) {
	pageSize := uint64(os.Getpagesize())
	var sqEntries, cqEntries uint32
	var ringMem, sqesMem uint64
	var memUsed uint64
	var ptr unsafe.Pointer

	errno := getSqCqEntries(entries, p, &sqEntries, &cqEntries)
	if errno != nil {
		return 0, errno
	}

	sqesMem = uint64(sqEntries) * uint64(unsafe.Sizeof(SubmissionQueue{}))
	sqesMem = (sqesMem + pageSize - 1) &^ (pageSize - 1)
	ringMem = uint64(cqEntries) * uint64(unsafe.Sizeof(CompletionQueue{}))
	if p.flags&SetupCQE32 != 0 {
		ringMem *= 2
	}
	ringMem += uint64(sqEntries) * uint64(unsafe.Sizeof(uint32(0)))
	memUsed = sqesMem + ringMem
	memUsed = (memUsed + pageSize - 1) &^ (pageSize - 1)

	if buf == nil && (sqesMem > hugePageSize || ringMem > hugePageSize) {
		return 0, syscall.ENOMEM
	}

	if buf != nil {
		if memUsed > bufSize {
			return 0, syscall.ENOMEM
		}
		ptr = buf
	} else {
		var mapHugetlb int
		if sqesMem <= pageSize {
			bufSize = pageSize
		} else {
			bufSize = hugePageSize
			mapHugetlb = syscall.MAP_HUGETLB
		}
		var err error
		ptr, err = sysMmap(
			0, uintptr(bufSize),
			syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_SHARED|syscall.MAP_ANONYMOUS|mapHugetlb, -1, 0)
		if err != nil {
			return 0, err
		}
	}

	sq.sqes = (*SubmissionQueueEntry)(ptr)
	if memUsed <= bufSize {
		sq.ringPtr = unsafe.Pointer(uintptr(unsafe.Pointer(sq.sqes)) + uintptr(sqesMem))
		cq.ringSize = 0
		sq.ringSize = 0
	} else {
		var mapHugetlb int
		if ringMem <= pageSize {
			bufSize = pageSize
		} else {
			bufSize = hugePageSize
			mapHugetlb = syscall.MAP_HUGETLB
		}
		var err error
		ptr, err = sysMmap(
			0, uintptr(bufSize),
			syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_SHARED|syscall.MAP_ANONYMOUS|mapHugetlb, -1, 0)
		if err != nil {
			_ = sysMunmap(uintptr(unsafe.Pointer(sq.sqes)), 1)

			return 0, err
		}
		sq.ringPtr = ptr
		sq.ringSize = uint(bufSize)
		cq.ringSize = 0
	}

	cq.ringPtr = sq.ringPtr
	p.sqOff.userAddr = uint64(uintptr(unsafe.Pointer(sq.sqes)))
	p.cqOff.userAddr = uint64(uintptr(sq.ringPtr))

	return uint(memUsed), nil
}

// liburing: __io_uring_queue_init_params
func (ring *Ring) internalQueueInitParams(entries uint32, p *Params, buf unsafe.Pointer, bufSize uint64) error {
	var fd int
	var sqEntries, index uint32
	var err error

	if p.flags&SetupRegisteredFdOnly != 0 && p.flags&SetupNoMmap == 0 {
		return syscall.EINVAL
	}

	if p.flags&SetupNoMmap != 0 {
		_, err = allocHuge(entries, p, ring.sqRing, ring.cqRing, buf, bufSize)
		if err != nil {
			return err
		}
		if buf != nil {
			ring.intFlags |= IntFlagAppMem
		}
	}

	fdPtr, _, errno := syscall.Syscall(sysSetup, uintptr(entries), uintptr(unsafe.Pointer(p)), 0)
	if errno != 0 {
		if p.flags&SetupNoMmap != 0 && ring.intFlags&IntFlagAppMem == 0 {
			_ = sysMunmap(uintptr(unsafe.Pointer(ring.sqRing.sqes)), 1)
			UnmapRings(ring.sqRing, ring.cqRing)
		}

		return errno
	}
	fd = int(fdPtr)

	if p.flags&SetupNoMmap == 0 {
		err = ring.QueueMmap(fd, p)
		if err != nil {
			syscall.Close(fd)

			return err
		}
	} else {
		SetupRingPointers(p, ring.sqRing, ring.cqRing)
	}

	sqEntries = *ring.sqRing.ringEntries
	for index = 0; index < sqEntries; index++ {
		*(*uint32)(
			unsafe.Add(unsafe.Pointer(ring.sqRing.array),
				index*uint32(unsafe.Sizeof(uint32(0))))) = index
	}

	ring.features = p.features
	ring.flags = p.flags
	ring.enterRingFd = fd
	if p.flags&SetupRegisteredFdOnly != 0 {
		ring.ringFd = -1
		ring.intFlags |= IntFlagRegRing | IntFlagRegRegRing
	} else {
		ring.ringFd = fd
	}

	return nil
}

// liburing: io_uring_queue_init_mem
func (ring *Ring) QueueInitMem(entries uint32, p *Params, buf unsafe.Pointer, bufSize uint64) error {
	// should already be set...
	p.flags |= SetupNoMmap

	return ring.internalQueueInitParams(entries, p, buf, bufSize)
}

// liburing: io_uring_queue_init_params - https://manpages.debian.org/unstable/liburing-dev/io_uring_queue_init_params.3.en.html
func (ring *Ring) QueueInitParams(entries uint32, p *Params) error {
	return ring.internalQueueInitParams(entries, p, nil, 0)
}

// liburing: io_uring_queue_init - https://manpages.debian.org/unstable/liburing-dev/io_uring_queue_init.3.en.html
func (ring *Ring) QueueInit(entries uint32, flags uint32) error {
	params := &Params{
		flags: flags,
	}

	return ring.QueueInitParams(entries, params)
}

// liburing: io_uring_queue_exit - https://manpages.debian.org/unstable/liburing-dev/io_uring_queue_exit.3.en.html
func (ring *Ring) QueueExit() {
	sq := ring.sqRing
	cq := ring.cqRing
	var sqeSize uintptr

	if sq.ringSize == 0 {
		sqeSize = unsafe.Sizeof(SubmissionQueueEntry{})
		if ring.flags&SetupSQE128 != 0 {
			sqeSize += 64
		}
		_ = sysMunmap(uintptr(unsafe.Pointer(sq.sqes)), sqeSize*uintptr(*sq.ringEntries))
		UnmapRings(sq, cq)
	} else if ring.intFlags&IntFlagAppMem == 0 {
		_ = sysMunmap(uintptr(unsafe.Pointer(sq.sqes)), uintptr(*sq.ringEntries)*unsafe.Sizeof(SubmissionQueueEntry{}))
		UnmapRings(sq, cq)
	}

	if ring.intFlags&IntFlagRegRing != 0 {
		_, _ = ring.UnregisterRingFd()
	}
	if ring.ringFd != -1 {
		syscall.Close(ring.ringFd)
	}
}

const ringSize = 320

func npages(size uint64, pageSize uint64) uint64 {
	size--
	size /= pageSize

	return uint64(fls(int(size)))
}

const (
	not63ul       = 18446744073709551552
	ringSizeCQOff = 63
)

// liburing: rings_size
func ringsSize(p *Params, entries uint32, cqEntries uint32, pageSize uint64) uint64 {
	var pages, sqSize, cqSize uint64

	cqSize = uint64(unsafe.Sizeof(CompletionQueueEvent{}))
	if p.flags&SetupCQE32 != 0 {
		cqSize += uint64(unsafe.Sizeof(CompletionQueueEvent{}))
	}
	cqSize *= uint64(cqEntries)
	cqSize += ringSize
	cqSize = (cqSize + ringSizeCQOff) & not63ul
	pages = 1 << npages(cqSize, pageSize)

	sqSize = uint64(unsafe.Sizeof(SubmissionQueueEntry{}))
	if p.flags&SetupSQE128 != 0 {
		sqSize += 64
	}
	sqSize *= uint64(entries)
	pages += 1 << npages(sqSize, pageSize)

	return pages * pageSize
}

// liburing: io_uring_mlock_size_params
func MlockSizeParams(entries uint32, p *Params) (uint64, error) {
	lp := &Params{}
	ring := NewRing()
	var cqEntries, sq uint32
	var pageSize uint64
	var err error

	err = ring.QueueInitParams(entries, lp)
	if err != nil {
		ring.QueueExit()
	}

	if lp.features&FeatNativeWorkers != 0 {
		return 0, nil
	}

	if entries == 0 {
		return 0, syscall.EINVAL
	}
	if entries > kernMaxEntries {
		if p.flags&SetupClamp == 0 {
			return 0, syscall.EINVAL
		}
		entries = kernMaxEntries
	}

	err = getSqCqEntries(entries, p, &sq, &cqEntries)
	if err != nil {
		return 0, err
	}

	pageSize = uint64(os.Getpagesize())

	return ringsSize(p, sq, cqEntries, pageSize), nil
}

// liburing: io_uring_mlock_size
func MlockSize(entries, flags uint32) (uint64, error) {
	p := &Params{}
	p.flags = flags

	return MlockSizeParams(entries, p)
}

// liburing: br_setup
func (ring *Ring) brSetup(nentries uint32, bgid uint16, flags uint32) (*BufAndRing, error) {
	var br *BufAndRing
	var reg BufReg
	var ringSize uintptr
	var brPtr unsafe.Pointer
	var err error

	reg = BufReg{}
	ringSize = uintptr(nentries) * unsafe.Sizeof(BufAndRing{})
	brPtr, err = sysMmap(
		0, ringSize, syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANONYMOUS|syscall.MAP_PRIVATE, -1, 0)
	if err != nil {
		return nil, err
	}
	br = (*BufAndRing)(brPtr)

	reg.RingAddr = uint64(uintptr(unsafe.Pointer(br)))
	reg.RingEntries = nentries
	reg.Bgid = bgid

	_, err = ring.RegisterBufferRing(&reg, flags)
	if err != nil {
		_ = sysMunmap(uintptr(unsafe.Pointer(br)), ringSize)

		return nil, err
	}

	return br, nil
}

// liburing: io_uring_setup_buf_ring - https://manpages.debian.org/unstable/liburing-dev/io_uring_setup_buf_ring.3.en.html
func (ring *Ring) SetupBufRing(nentries uint32, bgid int, flags uint32) (*BufAndRing, error) {
	br, err := ring.brSetup(nentries, uint16(bgid), flags)
	if br != nil {
		br.BufRingInit()
	}

	return br, err
}

// liburing: io_uring_free_buf_ring - https://manpages.debian.org/unstable/liburing-dev/io_uring_free_buf_ring.3.en.html
func (ring *Ring) FreeBufRing(bgid int) error {
	_, err := ring.UnregisterBufferRing(bgid)

	return err
}

func (ring *Ring) RingFd() int {
	return ring.ringFd
}
