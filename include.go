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

import "syscall"

// liburing: io_uring_sqe
type SubmissionQueueEntry struct {
	OpCode uint8
	Flags  uint8
	IoPrio uint16
	Fd     int32
	// union {
	// 	__u64	off;	/* offset into file */
	// 	__u64	addr2;
	// 	struct {
	// 		__u32	cmd_op;
	// 		__u32	__pad1;
	// 	};
	// };
	Off uint64
	// union {
	// 	__u64	addr;	/* pointer to buffer or iovecs */
	// 	__u64	splice_off_in;
	// };
	Addr uint64
	Len  uint32
	// union {
	// 	__kernel_rwf_t	rw_flags;
	// 	__u32		fsync_flags;
	// 	__u16		poll_events;	/* compatibility */
	// 	__u32		poll32_events;	/* word-reversed for BE */
	// 	__u32		sync_range_flags;
	// 	__u32		msg_flags;
	// 	__u32		timeout_flags;
	// 	__u32		accept_flags;
	// 	__u32		cancel_flags;
	// 	__u32		open_flags;
	// 	__u32		statx_flags;
	// 	__u32		fadvise_advice;
	// 	__u32		splice_flags;
	// 	__u32		rename_flags;
	// 	__u32		unlink_flags;
	// 	__u32		hardlink_flags;
	// 	__u32		xattr_flags;
	// 	__u32		msg_ring_flags;
	// 	__u32		uring_cmd_flags;
	// };
	OpcodeFlags uint32
	UserData    uint64
	// union {
	// 	/* index into fixed buffers, if used */
	// 	__u16	buf_index;
	// 	/* for grouped buffer selection */
	// 	__u16	buf_group;
	// } __attribute__((packed));
	BufIG       uint16
	Personality uint16
	// union {
	// 	__s32	splice_fd_in;
	// 	__u32	file_index;
	// 	struct {
	// 		__u16	addr_len;
	// 		__u16	__pad3[1];
	// 	};
	// };
	SpliceFdIn int32
	Addr3      uint64
	_pad2      [1]uint64
	// TODO: add __u8	cmd[0];
}

const FileIndexAlloc uint32 = 4294967295

const (
	SqeFixedFile uint8 = 1 << iota
	SqeIODrain
	SqeIOLink
	SqeIOHardlink
	SqeAsync
	SqeBufferSelect
	SqeCQESkipSuccess
)

const (
	SetupIOPoll uint32 = 1 << iota
	SetupSQPoll
	SetupSQAff
	SetupCQSize
	SetupClamp
	SetupAttachWQ
	SetupRDisabled
	SetupSubmitAll
	SetupCoopTaskrun
	SetupTaskrunFlag
	SetupSQE128
	SetupCQE32
	SetupSingleIssuer
	SetupDeferTaskrun
	SetupNoMmap
	SetupRegisteredFdOnly
)

const (
	OpNop uint8 = iota
	OpReadv
	OpWritev
	OpFsync
	OpReadFixed
	OpWriteFixed
	OpPollAdd
	OpPollRemove
	OpSyncFileRange
	OpSendmsg
	OpRecvmsg
	OpTimeout
	OpTimeoutRemove
	OpAccept
	OpAsyncCancel
	OpLinkTimeout
	OpConnect
	OpFallocate
	OpOpenat
	OpClose
	OpFilesUpdate
	OpStatx
	OpRead
	OpWrite
	OpFadvise
	OpMadvise
	OpSend
	OpRecv
	OpOpenat2
	OpEpollCtl
	OpSplice
	OpProvideBuffers
	OpRemoveBuffers
	OpTee
	OpShutdown
	OpRenameat
	OpUnlinkat
	OpMkdirat
	OpSymlinkat
	OpLinkat
	OpMsgRing
	OpFsetxattr
	OpSetxattr
	OpFgetxattr
	OpGetxattr
	OpSocket
	OpUringCmd
	OpSendZC
	OpSendMsgZC

	OpLast
)

const UringCmdFixed uint32 = 1 << 0

const FsyncDatasync uint32 = 1 << 0

const (
	TimeoutAbs uint32 = 1 << iota
	TimeoutUpdate
	TimeoutBoottime
	TimeoutRealtime
	LinkTimeoutUpdate
	TimeoutETimeSuccess
	TimeoutClockMask  = TimeoutBoottime | TimeoutRealtime
	TimeoutUpdateMask = TimeoutUpdate | LinkTimeoutUpdate
)

const SpliceFFdInFixed uint32 = 1 << 31

const (
	PollAddMulti uint32 = 1 << iota
	PollUpdateEvents
	PollUpdateUserData
	PollAddLevel
)

const (
	AsyncCancelAll uint32 = 1 << iota
	AsyncCancelFd
	AsyncCancelAny
	AsyncCancelFdFixed
)

const (
	RecvsendPollFirst uint16 = 1 << iota
	RecvMultishot
	RecvsendFixedBuf
	SendZCReportUsage
)

const NotifUsageZCCopied uint32 = 1 << 31

const (
	AcceptMultishot uint16 = 1 << iota
)

const (
	MsgData uint32 = iota
	MsgSendFd
)

const (
	MsgRingCQESkip uint32 = 1 << iota
	MsgRingFlagsPass
)

// liburing: io_uring_cqe
type CompletionQueueEvent struct {
	UserData uint64
	Res      int32
	Flags    uint32

	// FIXME
	// 	__u64 big_cqe[];
}

const (
	CQEFBuffer uint32 = 1 << iota
	CQEFMore
	CQEFSockNonempty
	CQEFNotif
)

const CQEBufferShift uint32 = 16

// Magic offsets for the application to mmap the data it needs.
const (
	offsqRing    uint64 = 0
	offcqRing    uint64 = 0x8000000
	offSQEs      uint64 = 0x10000000
	offPbufRing  uint64 = 0x80000000
	offPbufShift uint64 = 16
	offMmapMask  uint64 = 0xf8000000
)

// liburing: io_sqring_offsets
type SQRingOffsets struct {
	head        uint32
	tail        uint32
	ringMask    uint32
	ringEntries uint32
	flags       uint32
	dropped     uint32
	array       uint32
	resv1       uint32
	userAddr    uint64
}

const (
	SQNeedWakeup uint32 = 1 << iota
	SQCQOverflow
	SQTaskrun
)

// liburing: io_cqring_offsets
type CQRingOffsets struct {
	head        uint32
	tail        uint32
	ringMask    uint32
	ringEntries uint32
	overflow    uint32
	cqes        uint32
	flags       uint32
	resv1       uint32
	userAddr    uint64
}

const CQEventFdDisabled uint32 = 1 << 0

const (
	EnterGetEvents uint32 = 1 << iota
	EnterSQWakeup
	EnterSQWait
	EnterExtArg
	EnterRegisteredRing
)

// liburing: io_uring_params
type Params struct {
	sqEntries    uint32
	cqEntries    uint32
	flags        uint32
	sqThreadCPU  uint32
	sqThreadIdle uint32
	features     uint32
	wqFd         uint32
	resv         [3]uint32

	sqOff SQRingOffsets
	cqOff CQRingOffsets
}

const (
	FeatSingleMMap uint32 = 1 << iota
	FeatNoDrop
	FeatSubmitStable
	FeatRWCurPos
	FeatCurPersonality
	FeatFastPoll
	FeatPoll32Bits
	FeatSQPollNonfixed
	FeatExtArg
	FeatNativeWorkers
	FeatRcrcTags
	FeatCQESkip
	FeatLinkedFile
	FeatRegRegRing
)

const (
	RegisterBuffers uint32 = iota
	UnregisterBuffers

	RegisterFiles
	UnregisterFiles

	RegisterEventFD
	UnregisterEventFD

	RegisterFilesUpdate
	RegisterEventFDAsync
	RegisterProbe

	RegisterPersonality
	UnregisterPersonality

	RegisterRestrictions
	RegisterEnableRings

	RegisterFiles2
	RegisterFilesUpdate2
	RegisterBuffers2
	RegisterBuffersUpdate

	RegisterIOWQAff
	UnregisterIOWQAff

	RegisterIOWQMaxWorkers

	RegisterRingFDs
	UnregisterRingFDs

	RegisterPbufRing
	UnregisterPbufRing

	RegisterSyncCancel

	RegisterFileAllocRange

	RegisterLast

	RegisterUseRegisteredRing = 1 << 31
)

const (
	IOWQBound uint = iota
	IOWQUnbound
)

// liburing: io_uring_files_update
type FilesUpdate struct {
	Offset uint32
	Resv   uint32
	Fds    uint64
}

const (
	RsrcRegisterSparse uint32 = 1 << iota
)

// liburing: io_uring_rsrc_register
type RsrcRegister struct {
	Nr    uint32
	Flags uint32
	Resv2 uint64
	Data  uint64
	Tags  uint64
}

// liburing: io_uring_rsrc_update
type RsrcUpdate struct {
	Offset uint32
	Resv   uint32
	Data   uint64
}

// liburing: io_uring_rsrc_update2
type RsrcUpdate2 struct {
	Offset uint32
	Resv   uint32
	Data   uint64
	Tags   uint64
	Nr     uint32
	Resv2  uint32
}

const RegisterFilesSkip int = -2

const opSupported uint16 = 1 << 0

// liburing: io_uring_probe_op
type ProbeOp struct {
	Op    uint8
	Res   uint8
	Flags uint16
	Res2  uint32
}

// liburing: io_uring_probe
type Probe struct {
	LastOp uint8
	OpsLen uint8
	Res    uint16
	Res2   [3]uint32
	Ops    [probeOpsSize]ProbeOp
}

// liburing: io_uring_restriction
type Restriction struct {
	OpCode uint16
	// union {
	// 	__u8 register_op; /* IORING_RESTRICTION_REGISTER_OP */
	// 	__u8 sqe_op;      /* IORING_RESTRICTION_SQE_OP */
	// 	__u8 sqe_flags;   /* IORING_RESTRICTION_SQE_FLAGS_* */
	// };
	OpFlags uint8
	Resv    uint8
	Resv2   [3]uint32
}

// liburing: io_uring_buf
// liburing: io_uring_buf_ring
type BufAndRing struct {
	Addr uint64
	Len  uint32
	Bid  uint16
	Tail uint16
}

const PbufRingMMap = 1

// liburing: io_uring_buf_reg
type BufReg struct {
	RingAddr    uint64
	RingEntries uint32
	Bgid        uint16
	Pad         uint16
	Resv        [3]uint64
}

const (
	RestrictionRegisterOp uint32 = iota
	RestrictionSQEOp
	RestrictionSQEFlagsAllowed
	RestrictionSQEFlagsRequired

	RestrictionLast
)

// liburing: io_uring_getevents_arg
type GetEventsArg struct {
	sigMask   uint64
	sigMaskSz uint32
	pad       uint32
	ts        uint64
}

// liburing: io_uring_sync_cancel_reg
type SyncCancelReg struct {
	Addr    uint64
	Fd      int32
	Flags   uint32
	Timeout syscall.Timespec
	Pad     [4]uint64
}

// liburing: io_uring_file_index_range
type FileIndexRange struct {
	Off  uint32
	Len  uint32
	Resv uint64
}

// liburing: io_uring_recvmsg_out
type RecvmsgOut struct {
	Namelen    uint32
	ControlLen uint32
	PayloadLen uint32
	Flags      uint32
}

const (
	SocketUringOpSiocinq = iota
	SocketUringOpSiocoutq
)
