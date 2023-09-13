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
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// liburing: __io_uring_set_target_fixed_file
func (entry *SubmissionQueueEntry) setTargetFixedFile(fileIndex uint32) {
	entry.SpliceFdIn = int32(fileIndex + 1)
}

// liburing: io_uring_prep_rw
func (entry *SubmissionQueueEntry) prepareRW(opcode uint8, fd int, addr uintptr, length uint32, offset uint64) {
	entry.OpCode = opcode
	entry.Flags = 0
	entry.IoPrio = 0
	entry.Fd = int32(fd)
	entry.Off = offset
	entry.Addr = uint64(addr)
	entry.Len = length
	entry.UserData = 0
	entry.BufIG = 0
	entry.Personality = 0
	entry.SpliceFdIn = 0
}

// liburing: io_uring_prep_accept - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_accept.3.en.html
func (entry *SubmissionQueueEntry) PrepareAccept(fd int, addr uintptr, addrLen uint64, flags uint32) {
	entry.prepareRW(OpAccept, fd, addr, 0, addrLen)
	entry.OpcodeFlags = flags
}

// liburing: io_uring_prep_accept_direct - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_accept_direct.3.en.html
func (entry *SubmissionQueueEntry) PrepareAcceptDirect(
	fd int, addr uintptr, addrLen uint64, flags uint32, fileIndex uint32,
) {
	entry.PrepareAccept(fd, addr, addrLen, flags)

	if fileIndex == FileIndexAlloc {
		fileIndex--
	}

	entry.setTargetFixedFile(fileIndex)
}

// liburing: io_uring_prep_cancel - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_cancel.3.en.html
func (entry *SubmissionQueueEntry) PrepareCancel(userData uintptr, flags int) {
	entry.PrepareCancel64(uint64(userData), flags)
}

// liburing: io_uring_prep_cancel64 - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_cancel64.3.en.html
func (entry *SubmissionQueueEntry) PrepareCancel64(userData uint64, flags int) {
	entry.prepareRW(OpAsyncCancel, -1, 0, 0, 0)
	entry.Addr = userData
	entry.OpcodeFlags = uint32(flags)
}

// liburing: io_uring_prep_cancel_fd
func (entry *SubmissionQueueEntry) PrepareCancelFd(fd int, flags uint32) {
	entry.prepareRW(OpAsyncCancel, fd, 0, 0, 0)
	entry.OpcodeFlags = flags | AsyncCancelFd
}

// liburing: io_uring_prep_close - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_close.3.en.html
func (entry *SubmissionQueueEntry) PrepareClose(fd int) {
	entry.prepareRW(OpClose, fd, 0, 0, 0)
}

// liburing: io_uring_prep_close_direct - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_close_direct.3.en.html
func (entry *SubmissionQueueEntry) PrepareCloseDirect(fileIndex uint32) {
	entry.PrepareClose(0)
	entry.setTargetFixedFile(fileIndex)
}

// liburing: io_uring_prep_connect - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_connect.3.en.html
func (entry *SubmissionQueueEntry) PrepareConnect(fd int, addr uintptr, addrLen uint64) {
	entry.prepareRW(OpConnect, fd, addr, 0, addrLen)
}

// io_uring_prep_fadvise - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_fadvise.3.en.html
func (entry *SubmissionQueueEntry) PrepareFadvise(fd int, offset uint64, length int, advise uint32) {
	entry.prepareRW(OpFadvise, fd, 0, uint32(length), offset)
	entry.OpcodeFlags = advise
}

// liburing: io_uring_prep_fallocate - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_fallocate.3.en.html
func (entry *SubmissionQueueEntry) PrepareFallocate(fd int, mode int, offset, length uint64) {
	entry.prepareRW(OpFallocate, fd, 0, uint32(mode), offset)
	entry.Addr = length
}

// liburing: io_uring_prep_fgetxattr - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_fgetxattr.3.en.html
func (entry *SubmissionQueueEntry) PrepareFgetxattr(fd int, name, value []byte) {
	entry.prepareRW(OpFgetxattr, fd, uintptr(unsafe.Pointer(&name)),
		uint32(len(value)), uint64(uintptr(unsafe.Pointer(&value))))
}

// liburing: io_uring_prep_files_update - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_files_update.3.en.html
func (entry *SubmissionQueueEntry) PrepareFilesUpdate(fds []int, offset int) {
	entry.prepareRW(OpFilesUpdate, -1, uintptr(unsafe.Pointer(&fds)), uint32(len(fds)), uint64(offset))
}

// liburing: io_uring_prep_fsetxattr - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_fsetxattr.3.en.html
func (entry *SubmissionQueueEntry) PrepareFsetxattr(fd int, name, value []byte, flags int) {
	entry.prepareRW(
		OpFsetxattr, fd, uintptr(unsafe.Pointer(&name)), uint32(len(value)), uint64(uintptr(unsafe.Pointer(&value))))
	entry.OpcodeFlags = uint32(flags)
}

// liburing: io_uring_prep_fsync - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_fsync.3.en.html
func (entry *SubmissionQueueEntry) PrepareFsync(fd int, flags uint32) {
	entry.prepareRW(OpFsync, fd, 0, 0, 0)
	entry.OpcodeFlags = flags
}

// liburing: io_uring_prep_getxattr - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_getxattr.3.en.html
func (entry *SubmissionQueueEntry) PrepareGetxattr(name, value, path []byte) {
	entry.prepareRW(OpGetxattr, 0, uintptr(unsafe.Pointer(&name)),
		uint32(len(value)), uint64(uintptr(unsafe.Pointer(&value))))
	entry.Addr3 = uint64(uintptr(unsafe.Pointer(&path)))
	entry.OpcodeFlags = 0
}

// liburing: io_uring_prep_link - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_link.3.en.html
func (entry *SubmissionQueueEntry) PrepareLink(oldPath, newPath []byte, flags int) {
	entry.PrepareLinkat(unix.AT_FDCWD, oldPath, unix.AT_FDCWD, newPath, flags)
}

// liburing: io_uring_prep_link_timeout - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_link_timeout.3.en.html
func (entry *SubmissionQueueEntry) PrepareLinkTimeout(duration time.Duration, flags uint32) {
	spec := syscall.NsecToTimespec(duration.Nanoseconds())
	entry.prepareRW(OpLinkTimeout, -1, uintptr(unsafe.Pointer(&spec)), 1, 0)
	entry.OpcodeFlags = flags
}

// liburing: io_uring_prep_linkat - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_linkat.3.en.html
func (entry *SubmissionQueueEntry) PrepareLinkat(oldFd int, oldPath []byte, newFd int, newPath []byte, flags int) {
	entry.prepareRW(OpLinkat, oldFd, uintptr(unsafe.Pointer(&oldPath)),
		uint32(newFd), uint64(uintptr(unsafe.Pointer(&newPath))))
	entry.OpcodeFlags = uint32(flags)
}

// liburing: io_uring_prep_madvise - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_madvise.3.en.html
func (entry *SubmissionQueueEntry) PrepareMadvise(addr uintptr, length uint, advice int) {
	entry.prepareRW(OpMadvise, -1, addr, uint32(length), 0)
	entry.OpcodeFlags = uint32(advice)
}

// liburing: io_uring_prep_mkdir - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_mkdir.3.en.html
func (entry *SubmissionQueueEntry) PrepareMkdir(path []byte, mode uint32) {
	entry.PrepareMkdirat(unix.AT_FDCWD, path, mode)
}

// liburing: io_uring_prep_mkdirat - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_mkdirat.3.en.html
func (entry *SubmissionQueueEntry) PrepareMkdirat(dfd int, path []byte, mode uint32) {
	entry.prepareRW(OpMkdirat, dfd, uintptr(unsafe.Pointer(&path)), mode, 0)
}

// liburing: io_uring_prep_msg_ring - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_msg_ring.3.en.html
func (entry *SubmissionQueueEntry) PrepareMsgRing(fd int, length uint32, data uint64, flags uint32) {
	entry.prepareRW(OpMsgRing, fd, 0, length, data)
	entry.OpcodeFlags = flags
}

// liburing: io_uring_prep_msg_ring_cqe_flags - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_msg_ring_cqe_flags.3.en.html
func (entry *SubmissionQueueEntry) PrepareMsgRingCqeFlags(fd int, length uint32, data uint64, flags, cqeFlags uint32) {
	entry.prepareRW(OpMsgRing, fd, 0, length, data)
	entry.OpcodeFlags = MsgRingFlagsPass | flags
	entry.SpliceFdIn = int32(cqeFlags)
}

// liburing: io_uring_prep_msg_ring_fd - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_msg_ring_fd.3.en.html
func (entry *SubmissionQueueEntry) PrepareMsgRingFd(fd int, sourceFd int, targetFd int, data uint64, flags uint32) {
	entry.prepareRW(OpMsgRing, fd, uintptr(unsafe.Pointer(&msgDataVar)), 0, data)
	entry.Addr3 = uint64(sourceFd)
	if uint32(targetFd) == FileIndexAlloc {
		targetFd--
	}
	entry.setTargetFixedFile(uint32(targetFd))
	entry.OpcodeFlags = flags
}

// liburing: io_uring_prep_msg_ring_fd_alloc - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_msg_ring_fd_alloc.3.en.html
func (entry *SubmissionQueueEntry) PrepareMsgRingFdAlloc(fd int, sourceFd int, data uint64, flags uint32) {
	entry.PrepareMsgRingFd(fd, sourceFd, int(FileIndexAlloc), data, flags)
}

// liburing: io_uring_prep_multishot_accept - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_multishot_accept.3.en.html
func (entry *SubmissionQueueEntry) PrepareMultishotAccept(fd int, addr uintptr, addrLen uint64, flags int) {
	entry.PrepareAccept(fd, addr, addrLen, uint32(flags))
	entry.IoPrio |= AcceptMultishot
}

// liburing: io_uring_prep_multishot_accept_direct - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_multishot_accept_direct.3.en.html
func (entry *SubmissionQueueEntry) PrepareMultishotAcceptDirect(fd int, addr uintptr, addrLen uint64, flags int) {
	entry.PrepareMultishotAccept(fd, addr, addrLen, flags)
	entry.setTargetFixedFile(FileIndexAlloc - 1)
}

// liburing: io_uring_prep_nop - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_nop.3.en.html
func (entry *SubmissionQueueEntry) PrepareNop() {
	entry.prepareRW(OpNop, -1, 0, 0, 0)
}

// liburing: io_uring_prep_openat - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_openat.3.en.html
func (entry *SubmissionQueueEntry) PrepareOpenat(dfd int, path []byte, flags int, mode uint32) {
	entry.prepareRW(OpOpenat, dfd, uintptr(unsafe.Pointer(&path)), mode, 0)
	entry.OpcodeFlags = uint32(flags)
}

// liburing: io_uring_prep_openat2 - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_openat2.3.en.html
func (entry *SubmissionQueueEntry) PrepareOpenat2(dfd int, path []byte, openHow *unix.OpenHow) {
	entry.prepareRW(OpOpenat, dfd, uintptr(unsafe.Pointer(&path)),
		uint32(unsafe.Sizeof(*openHow)), uint64(uintptr(unsafe.Pointer(openHow))))
}

// liburing: io_uring_prep_openat2_direct - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_openat2_direct.3.en.html
func (entry *SubmissionQueueEntry) PrepareOpenat2Direct(dfd int, path []byte, openHow *unix.OpenHow, fileIndex uint32) {
	entry.PrepareOpenat2(dfd, path, openHow)
	if fileIndex == FileIndexAlloc {
		fileIndex--
	}
	entry.setTargetFixedFile(fileIndex)
}

// liburing: io_uring_prep_openat_direct - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_openat_direct.3.en.html
func (entry *SubmissionQueueEntry) PrepareOpenatDirect(dfd int, path []byte, flags int, mode uint32, fileIndex uint32) {
	entry.PrepareOpenat(dfd, path, flags, mode)
	if fileIndex == FileIndexAlloc {
		fileIndex--
	}
	entry.setTargetFixedFile(fileIndex)
}

// liburing: io_uring_prep_poll_add - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_poll_add.3.en.html
func (entry *SubmissionQueueEntry) PreparePollAdd(fd int, pollMask uint32) {
	entry.prepareRW(OpPollAdd, fd, 0, 0, 0)
	entry.OpcodeFlags = pollMask
}

// liburing: io_uring_prep_poll_multishot - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_poll_multishot.3.en.html
func (entry *SubmissionQueueEntry) PreparePollMultishot(fd int, pollMask uint32) {
	entry.PreparePollAdd(fd, pollMask)
	entry.Len = PollAddMulti
}

// liburing: io_uring_prep_poll_remove - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_poll_remove.3.en.html
func (entry *SubmissionQueueEntry) PreparePollRemove(userData uint64) {
	entry.prepareRW(OpPollRemove, -1, 0, 0, 0)
	entry.Addr = userData
}

// liburing: io_uring_prep_poll_update - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_poll_update.3.en.html
func (entry *SubmissionQueueEntry) PreparePollUpdate(oldUserData, newUserData uint64, pollMask, flags uint32) {
	entry.prepareRW(OpPollRemove, -1, 0, flags, newUserData)
	entry.Addr = oldUserData
	entry.OpcodeFlags = pollMask
}

// liburing: io_uring_prep_provide_buffers - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_provide_buffers.3.en.html
func (entry *SubmissionQueueEntry) PrepareProvideBuffers(addr uintptr, length, nr, bgid, bid int) {
	entry.prepareRW(OpProvideBuffers, nr, addr, uint32(length), uint64(bid))
	entry.BufIG = uint16(bgid)
}

// liburing: io_uring_prep_read - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_read.3.en.html
func (entry *SubmissionQueueEntry) PrepareRead(fd int, buf uintptr, nbytes uint32, offset uint64) {
	entry.prepareRW(OpRead, fd, buf, nbytes, offset)
}

// liburing: io_uring_prep_read_fixed - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_read_fixed.3.en.html
func (entry *SubmissionQueueEntry) PrepareReadFixed(
	fd int,
	buf uintptr,
	nbytes uint32,
	offset uint64,
	bufIndex int,
) {
	entry.prepareRW(OpReadFixed, fd, buf, nbytes, offset)
	entry.BufIG = uint16(bufIndex)
}

// liburing: io_uring_prep_readv - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_readv.3.en.html
func (entry *SubmissionQueueEntry) PrepareReadv(fd int, iovecs uintptr, nrVecs uint32, offset uint64) {
	entry.prepareRW(OpReadv, fd, iovecs, nrVecs, offset)
}

// liburing: io_uring_prep_readv2 - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_readv2.3.en.html
func (entry *SubmissionQueueEntry) PrepareReadv2(fd int, iovecs uintptr, nrVecs uint32, offset uint64, flags int) {
	entry.PrepareReadv(fd, iovecs, nrVecs, offset)
	entry.OpcodeFlags = uint32(flags)
}

// liburing: io_uring_prep_recv - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_recv.3.en.html
func (entry *SubmissionQueueEntry) PrepareRecv(
	fd int,
	buf uintptr,
	length uint32,
	flags int,
) {
	entry.prepareRW(OpRecv, fd, buf, length, 0)
	entry.OpcodeFlags = uint32(flags)
}

// liburing: io_uring_prep_recv_multishot - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_recv_multishot.3.en.html
func (entry *SubmissionQueueEntry) PrepareRecvMultishot(
	fd int,
	addr uintptr,
	length uint32,
	flags int,
) {
	entry.PrepareRecv(fd, addr, length, flags)
	entry.IoPrio |= RecvMultishot
}

// liburing: io_uring_prep_recvmsg - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_recvmsg.3.en.html
func (entry *SubmissionQueueEntry) PrepareRecvMsg(
	fd int,
	msg *syscall.Msghdr,
	flags uint32,
) {
	entry.prepareRW(OpRecvmsg, fd, uintptr(unsafe.Pointer(msg)), 1, 0)
	entry.OpcodeFlags = flags
}

// liburing: io_uring_prep_recvmsg_multishot - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_recvmsg_multishot.3.en.html
func (entry *SubmissionQueueEntry) PrepareRecvMsgMultishot(
	fd int,
	msg *syscall.Msghdr,
	flags uint32,
) {
	entry.PrepareRecvMsg(fd, msg, flags)
	entry.IoPrio |= RecvMultishot
}

// liburing: io_uring_prep_remove_buffers - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_remove_buffers.3.en.html
func (entry *SubmissionQueueEntry) PrepareRemoveBuffers(nr int, bgid int) {
	entry.prepareRW(OpRemoveBuffers, nr, 0, 0, 0)
	entry.BufIG = uint16(bgid)
}

// liburing: io_uring_prep_rename - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_rename.3.en.html
func (entry *SubmissionQueueEntry) PrepareRename(oldPath, netPath []byte) {
	entry.PrepareRenameat(unix.AT_FDCWD, oldPath, unix.AT_FDCWD, netPath, 0)
}

// liburing: io_uring_prep_renameat - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_renameat.3.en.html
func (entry *SubmissionQueueEntry) PrepareRenameat(
	oldFd int, oldPath []byte, newFd int, newPath []byte, flags uint32,
) {
	entry.prepareRW(OpRenameat, oldFd,
		uintptr(unsafe.Pointer(&oldPath)), uint32(newFd), uint64(uintptr(unsafe.Pointer(&newPath))))
	entry.OpcodeFlags = flags
}

// liburing: io_uring_prep_send - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_send.3.en.html
func (entry *SubmissionQueueEntry) PrepareSend(
	fd int,
	addr uintptr,
	length uint32,
	flags int,
) {
	entry.prepareRW(OpSend, fd, addr, length, 0)
	entry.OpcodeFlags = uint32(flags)
}

// liburing: io_uring_prep_send_set_addr - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_send_set_addr.3.en.html
func (entry *SubmissionQueueEntry) PrepareSendSetAddr(destAddr *syscall.Sockaddr, addrLen uint16) {
	entry.Off = uint64(uintptr(unsafe.Pointer(destAddr)))
	// FIXME?
	entry.SpliceFdIn = int32(addrLen)
}

// liburing: io_uring_prep_send_zc - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_send_zc.3.en.html
func (entry *SubmissionQueueEntry) PrepareSendZC(sockFd int, buf []byte, flags int, zcFlags uint32) {
	entry.prepareRW(OpSendZC, sockFd, uintptr(unsafe.Pointer(&buf)), uint32(len(buf)), 0)
	entry.OpcodeFlags = uint32(flags)
	entry.IoPrio = uint16(zcFlags)
}

// liburing: io_uring_prep_send_zc_fixed - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_send_zc_fixed.3.en.html
func (entry *SubmissionQueueEntry) PrepareSendZCFixed(sockFd int, buf []byte, flags int, zcFlags, bufIndex uint32) {
	entry.PrepareSendZC(sockFd, buf, flags, zcFlags)
	entry.IoPrio |= RecvsendFixedBuf
	entry.BufIG = uint16(bufIndex)
}

// liburing: io_uring_prep_sendmsg - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_sendmsg.3.en.html
func (entry *SubmissionQueueEntry) PrepareSendMsg(
	fd int,
	msg *syscall.Msghdr,
	flags uint32,
) {
	entry.prepareRW(OpSendmsg, fd, uintptr(unsafe.Pointer(msg)), 1, 0)
	entry.OpcodeFlags = flags
}

// liburing: io_uring_prep_sendmsg_zc - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_sendmsg_zc.3.en.html
func (entry *SubmissionQueueEntry) PrepareSendmsgZC(fd int, msg *syscall.Msghdr, flags uint32) {
	entry.PrepareSendMsg(fd, msg, flags)
	entry.OpCode = OpSendMsgZC
}

// liburing: io_uring_prep_sendto - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_sendto.3.en.html
func (entry *SubmissionQueueEntry) PrepareSendto(
	sockFd int, buf []byte, flags int, addr *syscall.Sockaddr, addrLen uint32,
) {
	entry.PrepareSend(sockFd, uintptr(unsafe.Pointer(&buf)), uint32(len(buf)), flags)
	entry.PrepareSendSetAddr(addr, uint16(addrLen))
}

// liburing: io_uring_prep_setxattr - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_setxattr.3.en.html
func (entry *SubmissionQueueEntry) PrepareSetxattr(name, value, path []byte, flags int, length uint32) {
	entry.prepareRW(OpSetxattr, 0, uintptr(unsafe.Pointer(&name)), length, uint64(uintptr(unsafe.Pointer(&value))))
	entry.Addr3 = uint64(uintptr(unsafe.Pointer(&path)))
	entry.OpcodeFlags = uint32(flags)
}

// liburing: io_uring_prep_shutdown - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_shutdown.3.en.html
func (entry *SubmissionQueueEntry) PrepareShutdown(fd, how int) {
	entry.prepareRW(OpShutdown, fd, 0, uint32(how), 0)
}

// liburing: io_uring_prep_socket - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_socket.3.en.html
func (entry *SubmissionQueueEntry) PrepareSocket(domain, socketType, protocol int, flags uint32) {
	entry.prepareRW(OpSocket, domain, 0, uint32(protocol), uint64(socketType))
	entry.OpcodeFlags = flags
}

// liburing: io_uring_prep_socket_direct - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_socket_direct.3.en.html
func (entry *SubmissionQueueEntry) PrepareSocketDirect(domain, socketType, protocol int, fileIndex, flags uint32) {
	entry.PrepareSocket(domain, socketType, protocol, flags)
	if fileIndex == FileIndexAlloc {
		fileIndex--
	}
	entry.setTargetFixedFile(fileIndex)
}

// liburing: io_uring_prep_socket_direct_alloc - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_socket_direct_alloc.3.en.html
func (entry *SubmissionQueueEntry) PrepareSocketDirectAlloc(domain, socketType, protocol int, flags uint32) {
	entry.PrepareSocket(domain, socketType, protocol, flags)
	entry.setTargetFixedFile(FileIndexAlloc - 1)
}

// liburing: io_uring_prep_splice - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_splice.3.en.html
func (entry *SubmissionQueueEntry) PrepareSplice(
	fdIn int, offIn int64, fdOut int, offOut int64, nbytes, spliceFlags uint32,
) {
	entry.prepareRW(OpSplice, fdOut, 0, nbytes, uint64(offOut))
	entry.Addr = uint64(offIn)
	entry.SpliceFdIn = int32(fdIn)
	entry.OpcodeFlags = spliceFlags
}

// liburing: io_uring_prep_statx - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_statx.3.en.html
func (entry *SubmissionQueueEntry) PrepareStatx(dfd int, path []byte, flags int, mask uint32, statx *unix.Statx_t) {
	entry.prepareRW(OpStatx, dfd, uintptr(unsafe.Pointer(&path)), mask, uint64(uintptr(unsafe.Pointer(statx))))
	entry.OpcodeFlags = uint32(flags)
}

// liburing: io_uring_prep_symlink - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_symlink.3.en.html
func (entry *SubmissionQueueEntry) PrepareSymlink(target, linkpath []byte) {
	entry.PrepareSymlinkat(target, unix.AT_FDCWD, linkpath)
}

// liburing: io_uring_prep_symlinkat - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_symlinkat.3.en.html
func (entry *SubmissionQueueEntry) PrepareSymlinkat(target []byte, newdirfd int, linkpath []byte) {
	entry.prepareRW(OpSymlinkat, newdirfd, uintptr(unsafe.Pointer(&target)), 0, uint64(uintptr(unsafe.Pointer(&linkpath))))
}

// liburing: io_uring_prep_sync_file_range - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_sync_file_range.3.en.html
func (entry *SubmissionQueueEntry) PrepareSyncFileRange(fd int, length uint32, offset uint64, flags int) {
	entry.prepareRW(OpSyncFileRange, fd, 0, length, offset)
	entry.OpcodeFlags = uint32(flags)
}

// liburing: io_uring_prep_tee - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_tee.3.en.html
func (entry *SubmissionQueueEntry) PrepareTee(fdIn, fdOut int, nbytes, spliceFlags uint32) {
	entry.prepareRW(OpTee, fdOut, 0, nbytes, 0)
	entry.Addr = 0
	entry.SpliceFdIn = int32(fdIn)
	entry.OpcodeFlags = spliceFlags
}

// liburing: io_uring_prep_timeout - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_timeout.3.en.html
func (entry *SubmissionQueueEntry) PrepareTimeout(spec *syscall.Timespec, count, flags uint32) {
	entry.prepareRW(OpTimeout, -1, uintptr(unsafe.Pointer(&spec)), 1, uint64(count))
	entry.OpcodeFlags = flags
}

// liburing: io_uring_prep_timeout_remove - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_timeout_remove.3.en.html
func (entry *SubmissionQueueEntry) PrepareTimeoutRemove(duration time.Duration, count uint64, flags uint32) {
	spec := syscall.NsecToTimespec(duration.Nanoseconds())
	entry.prepareRW(OpTimeoutRemove, -1, uintptr(unsafe.Pointer(&spec)), 1, count)
	entry.OpcodeFlags = flags
}

// liburing: io_uring_prep_timeout_update - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_timeout_update.3.en.html
func (entry *SubmissionQueueEntry) PrepareTimeoutUpdate(duration time.Duration, count uint64, flags uint32) {
	spec := syscall.NsecToTimespec(duration.Nanoseconds())
	entry.prepareRW(OpTimeoutRemove, -1, uintptr(unsafe.Pointer(&spec)), 1, count)
	entry.OpcodeFlags = flags | TimeoutUpdate
}

// liburing: io_uring_prep_unlink - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_unlink.3.en.html
func (entry *SubmissionQueueEntry) PrepareUnlink(path uintptr, flags int) {
	entry.PrepareUnlinkat(unix.AT_FDCWD, path, flags)
}

// liburing: io_uring_prep_unlinkat - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_unlinkat.3.en.html
func (entry *SubmissionQueueEntry) PrepareUnlinkat(dfd int, path uintptr, flags int) {
	entry.prepareRW(OpUnlinkat, dfd, path, 0, 0)
	entry.OpcodeFlags = uint32(flags)
}

// liburing: io_uring_prep_write - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_write.3.en.html
func (entry *SubmissionQueueEntry) PrepareWrite(fd int, buf uintptr, nbytes uint32, offset uint64) {
	entry.prepareRW(OpWrite, fd, buf, nbytes, offset)
}

// liburing: io_uring_prep_write_fixed - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_write_fixed.3.en.html
func (entry *SubmissionQueueEntry) PrepareWriteFixed(
	fd int,
	vectors uintptr,
	length uint32,
	offset uint64,
	index int,
) {
	entry.prepareRW(OpWriteFixed, fd, vectors, length, offset)
	entry.BufIG = uint16(index)
}

// liburing: io_uring_prep_writev - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_writev.3.en.html
func (entry *SubmissionQueueEntry) PrepareWritev(
	fd int,
	iovecs uintptr,
	nrVecs uint32,
	offset uint64,
) {
	entry.prepareRW(OpWritev, fd, iovecs, nrVecs, offset)
}

// liburing: io_uring_prep_writev2 - https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_writev2.3.en.html
func (entry *SubmissionQueueEntry) PrepareWritev2(
	fd int,
	iovecs uintptr,
	nrVecs uint32,
	offset uint64,
	flags int,
) {
	entry.PrepareWritev(fd, iovecs, nrVecs, offset)
	entry.OpcodeFlags = uint32(flags)
}

const bit32Offset = 32

// liburing: io_uring_prep_cmd_sock
func (entry *SubmissionQueueEntry) PrepareCmdSock(
	cmdOp int, fd int, _ int, _ int, _ unsafe.Pointer, _ int,
) {
	// This will be removed once the get/setsockopt() patches land
	// var unused uintptr
	// unused = uintptr(optlen)
	// unused = uintptr(optval)
	// unused = uintptr(level)
	// unused = uintptr(optname)
	// io_uring_prep_rw(IORING_OP_URING_CMD, sqe, fd, nil, 0, 0)
	entry.prepareRW(OpUringCmd, fd, 0, 0, 0)
	entry.Off = uint64(cmdOp << bit32Offset)
}
