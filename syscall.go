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
	"unsafe"
)

/*
 * io_uring_enter - https://manpages.debian.org/unstable/liburing-dev/io_uring_enter.2.en.html
 */
func (ring *Ring) Enter(submitted uint32, waitNr uint32, flags uint32, sig unsafe.Pointer) (uint, error) {
	return ring.Enter2(submitted, waitNr, flags, sig, nSig/szDivider)
}

/*
 * io_uring_enter2 - https://manpages.debian.org/unstable/liburing-dev/io_uring_enter.2.en.html
 */
func (ring *Ring) Enter2(
	submitted uint32,
	waitNr uint32,
	flags uint32,
	sig unsafe.Pointer,
	size int,
) (uint, error) {
	var (
		consumed uintptr
		errno    syscall.Errno
	)

	consumed, _, errno = syscall.Syscall6(
		sysEnter,
		uintptr(ring.enterRingFd),
		uintptr(submitted),
		uintptr(waitNr),
		uintptr(flags),
		uintptr(sig),
		uintptr(size),
	)

	if errno > 0 {
		return 0, errno
	}

	return uint(consumed), nil
}

// liburing: io_uring_setup - https://manpages.debian.org/unstable/liburing-dev/io_uring_setup.2.en.html
func Setup(entries uint32, p *Params) (uint, error) {
	fd, _, errno := syscall.Syscall(sysSetup, uintptr(entries), uintptr(unsafe.Pointer(p)), 0)
	runtime.KeepAlive(p)

	return uint(fd), errno
}

func syscallRegister(fd int, opcode uint32, arg unsafe.Pointer, nrArgs uint32) (uint, syscall.Errno) {
	returnFirst, _, errno := syscall.Syscall6(
		sysRegister,
		uintptr(fd),
		uintptr(opcode),
		uintptr(arg),
		uintptr(nrArgs),
		0,
		0,
	)

	return uint(returnFirst), errno
}

// liburing: io_uring_register - https://manpages.debian.org/unstable/liburing-dev/io_uring_register.2.en.html
func (ring *Ring) Register(fd int, opcode uint32, arg unsafe.Pointer, nrArgs uint32) (uint, syscall.Errno) {
	return syscallRegister(fd, opcode, arg, nrArgs)
}
