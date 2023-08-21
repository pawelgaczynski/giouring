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

func TestIsOpSupported(t *testing.T) {
	for _, opCode := range []uint8{
		OpNop,
		OpReadv,
		OpWritev,
		OpFsync,
		OpReadFixed,
		OpWriteFixed,
		OpPollAdd,
		OpPollRemove,
		OpSyncFileRange,
		OpSendmsg,
		OpRecvmsg,
		OpTimeout,
		OpTimeoutRemove,
		OpAccept,
		OpAsyncCancel,
		OpLinkTimeout,
		OpConnect,
		OpFallocate,
		OpOpenat,
		OpClose,
		OpFilesUpdate,
		OpStatx,
		OpRead,
		OpWrite,
		OpFadvise,
		OpMadvise,
		OpSend,
		OpRecv,
		OpOpenat2,
		OpEpollCtl,
		OpSplice,
		OpProvideBuffers,
		OpRemoveBuffers,
		OpTee,
		OpShutdown,
		OpRenameat,
		OpUnlinkat,
		OpMkdirat,
		OpSymlinkat,
		OpLinkat,
		OpMsgRing,
		OpFsetxattr,
		OpSetxattr,
		OpFgetxattr,
		OpGetxattr,
		OpSocket,
		OpUringCmd,
		OpSendZC,
		OpSendMsgZC,
	} {
		probe, err := GetProbe()
		NoError(t, err)
		NotNil(t, probe)
		supported := probe.IsSupported(opCode)
		Equal(t, true, supported)
	}
}
