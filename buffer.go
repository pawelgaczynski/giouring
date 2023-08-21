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
	"sync/atomic"
	"unsafe"
)

var RingBufStructSize = uint16(unsafe.Sizeof(BufAndRing{}))

// liburing: io_uring_buf_ring_add - https://manpages.debian.org/unstable/liburing-dev/io_uring_buf_ring_add.3.en.html
func (br *BufAndRing) BufRingAdd(addr uintptr, length uint32, bid uint16, mask, bufOffset int) {
	buf := (*BufAndRing)(
		unsafe.Pointer(uintptr(unsafe.Pointer(br)) +
			(uintptr(((br.Tail + uint16(bufOffset)) & uint16(mask)) * RingBufStructSize))))
	buf.Addr = uint64(addr)
	buf.Len = length
	buf.Bid = bid
}

const bit16offset = 16

// liburing: io_uring_buf_ring_advance - https://manpages.debian.org/unstable/liburing-dev/io_uring_buf_ring_advance.3.en.html
func (br *BufAndRing) BufRingAdvance(count int) {
	newTail := br.Tail + uint16(count)
	// FIXME: implement 16 bit version of atomic.Store
	bidAndTail := (*uint32)(unsafe.Pointer(&br.Bid))
	bidAndTailVal := uint32(newTail)<<bit16offset + uint32(br.Bid)
	atomic.StoreUint32(bidAndTail, bidAndTailVal)
}

// liburing:  __io_uring_buf_ring_cq_advance - https://manpages.debian.org/unstable/liburing-dev/__io_uring_buf_ring_cq_advance.3.en.html
func (ring *Ring) internalBufRingCQAdvance(br *BufAndRing, bufCount, cqeCount int) {
	br.Tail += uint16(bufCount)
	ring.CQAdvance(uint32(cqeCount))
}

// liburing: io_uring_buf_ring_cq_advance - https://manpages.debian.org/unstable/liburing-dev/io_uring_buf_ring_cq_advance.3.en.html
func (ring *Ring) BufRingCQAdvance(br *BufAndRing, count int) {
	ring.internalBufRingCQAdvance(br, count, count)
}

// liburing: io_uring_buf_ring_init - https://manpages.debian.org/unstable/liburing-dev/io_uring_buf_ring_init.3.en.html
func (br *BufAndRing) BufRingInit() {
	br.Tail = 0
}

// liburing: io_uring_buf_ring_mask - https://manpages.debian.org/unstable/liburing-dev/io_uring_buf_ring_mask.3.en.html
func BufRingMask(entries uint32) int {
	return int(entries - 1)
}
