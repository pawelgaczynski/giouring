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
	"unsafe"

	. "github.com/stretchr/testify/require"
)

func TestPrepareMsgRing(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareMsgRing(10, 100, 200, 60)

	Equal(t, uint8(40), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(200), entry.Off)
	Equal(t, uint64(0), entry.Addr)
	Equal(t, uint32(100), entry.Len)
	Equal(t, uint32(60), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareAccept(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareAccept(10, 100, 200, 60)

	Equal(t, uint8(13), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(200), entry.Off)
	Equal(t, uint64(100), entry.Addr)
	Equal(t, uint32(0), entry.Len)
	Equal(t, uint32(60), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareClose(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareClose(10)

	Equal(t, uint8(19), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(0), entry.Off)
	Equal(t, uint64(0), entry.Addr)
	Equal(t, uint32(0), entry.Len)
	Equal(t, uint32(0), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareCloseDirect(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareCloseDirect(10)

	Equal(t, uint8(19), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(0), entry.Fd)
	Equal(t, uint64(0), entry.Off)
	Equal(t, uint64(0), entry.Addr)
	Equal(t, uint32(0), entry.Len)
	Equal(t, uint32(0), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(11), entry.SpliceFdIn)
}

func TestPrepareReadv(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareReadv(10, uintptr(12345), 60, 10)

	Equal(t, uint8(1), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(10), entry.Off)
	Equal(t, uint64(12345), entry.Addr)
	Equal(t, uint32(60), entry.Len)
	Equal(t, uint32(0), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareReadv2(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareReadv2(10, uintptr(12345), 60, 10, 15)

	Equal(t, uint8(1), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(10), entry.Off)
	Equal(t, uint64(12345), entry.Addr)
	Equal(t, uint32(60), entry.Len)
	Equal(t, uint32(15), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareReadFixed(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareReadFixed(10, uintptr(12345), 60, 10, 15)

	Equal(t, uint8(4), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(10), entry.Off)
	Equal(t, uint64(12345), entry.Addr)
	Equal(t, uint32(60), entry.Len)
	Equal(t, uint32(0), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(15), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareWritev(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareWritev(10, uintptr(12345), 60, 10)

	Equal(t, uint8(2), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(10), entry.Off)
	Equal(t, uint64(12345), entry.Addr)
	Equal(t, uint32(60), entry.Len)
	Equal(t, uint32(0), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareWritev2(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareWritev2(10, uintptr(12345), 60, 10, 15)

	Equal(t, uint8(2), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(10), entry.Off)
	Equal(t, uint64(12345), entry.Addr)
	Equal(t, uint32(60), entry.Len)
	Equal(t, uint32(15), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareWriteFixed(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareWriteFixed(10, uintptr(12345), 60, 10, 15)

	Equal(t, uint8(5), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(10), entry.Off)
	Equal(t, uint64(12345), entry.Addr)
	Equal(t, uint32(60), entry.Len)
	Equal(t, uint32(0), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(15), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareSendMsg(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareSend(10, uintptr(12345), 60, 10)

	Equal(t, uint8(26), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(0), entry.Off)
	Equal(t, uint64(12345), entry.Addr)
	Equal(t, uint32(60), entry.Len)
	Equal(t, uint32(10), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareNop(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareNop()

	Equal(t, uint8(0), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(-1), entry.Fd)
	Equal(t, uint64(0), entry.Off)
	Equal(t, uint64(0), entry.Addr)
	Equal(t, uint32(0), entry.Len)
	Equal(t, uint32(0), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareTimeout(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	duration := time.Second
	spec := syscall.NsecToTimespec(duration.Nanoseconds())
	entry.PrepareTimeout(&spec, 10, 15)

	Equal(t, uint8(11), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(-1), entry.Fd)
	Equal(t, uint64(10), entry.Off)
	NotZero(t, entry.Addr)
	Equal(t, uint32(1), entry.Len)
	Equal(t, uint32(15), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareTimeoutRemove(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	duration := time.Second
	entry.PrepareTimeoutRemove(duration, 10, 15)

	Equal(t, uint8(12), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(-1), entry.Fd)
	Equal(t, uint64(10), entry.Off)
	NotZero(t, entry.Addr)
	Equal(t, uint32(1), entry.Len)
	Equal(t, uint32(15), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareTimeoutUpdate(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	duration := time.Second
	entry.PrepareTimeoutUpdate(duration, 10, 15)

	Equal(t, uint8(12), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(-1), entry.Fd)
	Equal(t, uint64(10), entry.Off)
	NotZero(t, entry.Addr)
	Equal(t, uint32(1), entry.Len)
	Equal(t, uint32(15), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareAcceptDirect(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareAcceptDirect(10, uintptr(12345), 10, 15, 7)

	Equal(t, uint8(13), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(10), entry.Off)
	Equal(t, uint64(12345), entry.Addr)
	Equal(t, uint32(0), entry.Len)
	Equal(t, uint32(15), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(8), entry.SpliceFdIn)
}

func TestPrepareSend(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareSend(10, uintptr(12345), 10, 15)

	Equal(t, uint8(26), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(0), entry.Off)
	Equal(t, uint64(12345), entry.Addr)
	Equal(t, uint32(10), entry.Len)
	Equal(t, uint32(15), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareRecv(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareRecv(10, uintptr(12345), 10, 15)

	Equal(t, uint8(27), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(0), entry.Off)
	Equal(t, uint64(12345), entry.Addr)
	Equal(t, uint32(10), entry.Len)
	Equal(t, uint32(15), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareRecvMsg(t *testing.T) {
	var (
		msg   syscall.Msghdr
		entry = &SubmissionQueueEntry{}
	)
	msgAddr := uint64(uintptr(unsafe.Pointer(&msg)))

	entry.PrepareRecvMsg(10, &msg, 15)

	Equal(t, uint8(10), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(0), entry.Off)
	Equal(t, msgAddr, entry.Addr)
	Equal(t, uint32(1), entry.Len)
	Equal(t, uint32(15), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareRecvMultishot(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareRecvMultishot(10, uintptr(12345), 10, 15)

	Equal(t, uint8(27), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(2), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(0), entry.Off)
	Equal(t, uint64(12345), entry.Addr)
	Equal(t, uint32(10), entry.Len)
	Equal(t, uint32(15), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(0), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}

func TestPrepareProvideBuffers(t *testing.T) {
	entry := &SubmissionQueueEntry{}
	entry.PrepareProvideBuffers(uintptr(12345), 16, 10, 3, 10)

	Equal(t, uint8(31), entry.OpCode)
	Equal(t, uint8(0), entry.Flags)
	Equal(t, uint16(0), entry.IoPrio)
	Equal(t, int32(10), entry.Fd)
	Equal(t, uint64(10), entry.Off)
	Equal(t, uint64(12345), entry.Addr)
	Equal(t, uint32(16), entry.Len)
	Equal(t, uint32(0), entry.OpcodeFlags)
	Equal(t, uint64(0), entry.UserData)
	Equal(t, uint16(3), entry.BufIG)
	Equal(t, uint16(0), entry.Personality)
	Equal(t, int32(0), entry.SpliceFdIn)
}
