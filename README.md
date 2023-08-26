<a name="readme-top"></a>

# giouring - about the project

**giouring** is a Go port of the [liburing](https://github.com/axboe/liburing) library. It is written entirely in Go. No cgo.

Almost all functions and structures from [liburing](https://github.com/axboe/liburing) was implemented.

* **giouring** versioning is aligned with [liburing](https://github.com/axboe/liburing) versioning.
* **giouring** is currently up to date with [liburing](https://github.com/axboe/liburing) commit: [e1e758ae8360521334399c2a6eace05fa518e218](https://github.com/axboe/liburing/commit/e1e758ae8360521334399c2a6eace05fa518e218)


The **giouring** API is very similar to the [liburing](https://github.com/axboe/liburing) API, so anyone familiar with [liburing](https://github.com/axboe/liburing) will find it easier when writing code. Significant changes include:
* Method and structure names have been aligned with the naming conventions of the Go language.
* The prefix *io_uring* has been removed from method and structure names. After importing the package, methods and types will be preceded by the library name: *giouring*.
* *SQE* and *CQE* types have been given full names: *SubmissionQueueEntry* and *CompletionQueueEvent*.
* Additionally, if a method primarily pertains to a specific structure, for example, all methods prefixed with *io_uring_prep* that are related to the *SubmissionQueueEntry* structure (in [liburing](https://github.com/axboe/liburing): *io_uring_sqe*), the pointer that was passed in C as a method argument has been moved to the method receiver.

#### Important notice

* **giouring** was tested on kernel version *6.2.0-27-generic*. Keep in mind that when running unit tests on older kernel versions, some tests may fail because the older kernel may not support some functionality. This will be fixed in the future.
* Test coverage is currently low, but it will be systematically expanded.

### Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/pawelgaczynski/giouring.svg)](https://pkg.go.dev/github.com/pawelgaczynski/giouring)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Used by

* [Gain - a high-performance io_uring networking framework written entirely in Go.](https://github.com/pawelgaczynski/gain)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Prerequisites
Gain requires Go 1.20+

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Implemented structs

| liburing name | Golang liburing port name | Notes | Implemented |
| -------- | ------- | ------- | ------- |
| io_uring_sq | SubmissionQueue |  | :heavy_check_mark: |
| io_uring_cq | CompletionQueue |  | :heavy_check_mark: |
| io_uring | Ring |  | :heavy_check_mark: |
| io_uring_sqe | SubmissionQueueEntry |  | :heavy_check_mark: |
| io_uring_cqe | CompletionQueueEvent |  | :heavy_check_mark: |
| io_sqring_offsets | SQRingOffsets |  | :heavy_check_mark: |
| io_cqring_offsets | CQRingOffsets |  | :heavy_check_mark: |
| io_uring_params | Params |  | :heavy_check_mark: |
| io_uring_files_update | FilesUpdate |  | :heavy_check_mark: |
| io_uring_rsrc_register | RsrcRegister |  | :heavy_check_mark: |
| io_uring_rsrc_update | RsrcUpdate |  | :heavy_check_mark: |
| io_uring_rsrc_update2 | RsrcUpdate2 |  | :heavy_check_mark: |
| io_uring_probe_op | ProbeOp |  | :heavy_check_mark: |
| io_uring_probe | Probe |  | :heavy_check_mark: |
| io_uring_restriction | Restriction |  | :heavy_check_mark: |
| io_uring_buf | BufAndRing |  | :heavy_check_mark: |
| io_uring_buf_ring | BufAndRing |  | :heavy_check_mark: |
| io_uring_buf_reg | BufReg |  | :heavy_check_mark: |
| io_uring_getevents_arg | GetEventsArg |  | :heavy_check_mark: |
| io_uring_sync_cancel_reg | SyncCancelReg |  | :heavy_check_mark: |
| io_uring_file_index_range | FileIndexRange |  | :heavy_check_mark: |
| io_uring_recvmsg_out | RecvmsgOut |  | :heavy_check_mark: |

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Implemented methods

| liburing name | Receiver type | Golang liburing port name | Notes | Implemented |
| -------- | ------- | ------- | ------- | ------- |
| [IO_URING_CHECK_VERSION](https://manpages.debian.org/unstable/liburing-dev/IO_URING_CHECK_VERSION.3.en.html) |  |  |  |  |
| [IO_URING_VERSION_MAJOR](https://manpages.debian.org/unstable/liburing-dev/IO_URING_CHECK_VERSION.3.en.html) |  |  |  |  |
| [IO_URING_VERSION_MINOR](https://manpages.debian.org/unstable/liburing-dev/IO_URING_CHECK_VERSION.3.en.html) |  |  |  |  |
| [io_uring_buf_ring_add](https://manpages.debian.org/unstable/liburing-dev/io_uring_buf_ring_add.3.en.html) | BufAndRing | [BufRingAdd](buffer.go) | | :heavy_check_mark: |
| [io_uring_buf_ring_advance](https://manpages.debian.org/unstable/liburing-dev/io_uring_buf_ring_advance.3.en.html) | BufAndRing | [BufRingAdvance](buffer.go) |  | :heavy_check_mark: |
| [io_uring_buf_ring_cq_advance](https://manpages.debian.org/unstable/liburing-dev/io_uring_buf_ring_cq_advance.3.en.html) | Ring | [BufRingCQAdvance](buffer.go) |  | :heavy_check_mark: |
| [io_uring_buf_ring_init](https://manpages.debian.org/unstable/liburing-dev/io_uring_buf_ring_init.3.en.html) | BufAndRing | [BufRingInit](buffer.go) |  | :heavy_check_mark: |
| [io_uring_buf_ring_mask](https://manpages.debian.org/unstable/liburing-dev/io_uring_buf_ring_mask.3.en.html) |  | [BufRingMask]() |  | :heavy_check_mark: |
| [io_uring_check_version](https://manpages.debian.org/unstable/liburing-dev/io_uring_check_version.3.en.html) |  | [CheckVersion](version.go) |  | :heavy_check_mark: |
| [io_uring_close_ring_fd](https://manpages.debian.org/unstable/liburing-dev/io_uring_close_ring_fd.3.en.html) | Ring | [CloseRingFd](register.go) |  | :heavy_check_mark: |
| [io_uring_cq_advance](https://manpages.debian.org/unstable/liburing-dev/io_uring_cq_advance.3.en.html) | Ring | [CQAdvance](lib.go) |  | :heavy_check_mark: |
| [io_uring_cq_has_overflow](https://manpages.debian.org/unstable/liburing-dev/io_uring_cq_has_overflow.3.en.html) | Ring | [CQHasOverflow](lib.go) |  | :heavy_check_mark: |
| [io_uring_cq_ready](https://manpages.debian.org/unstable/liburing-dev/io_uring_cq_ready.3.en.html) | Ring | [CQReady](lib.go) |  | :heavy_check_mark: |
| [io_uring_cqe_get_data](https://manpages.debian.org/unstable/liburing-dev/io_uring_cqe_get_data.3.en.html) | CompletionQueueEvent | [GetData](lib.go) |  | :heavy_check_mark: |
| [io_uring_cqe_get_data64](https://manpages.debian.org/unstable/liburing-dev/io_uring_cqe_get_data64.3.en.html) |  CompletionQueueEvent| [GetData64](lib.go) |  | :heavy_check_mark: |
| [io_uring_cqe_seen](https://manpages.debian.org/unstable/liburing-dev/io_uring_cqe_seen.3.en.html) | Ring | [CQESeen](lib.go) |  | :heavy_check_mark: |
| [io_uring_enter](https://manpages.debian.org/unstable/liburing-dev/io_uring_enter.2.en.html) | Ring | [Enter](syscall.go) |  | :heavy_check_mark: |
| [io_uring_enter2](https://manpages.debian.org/unstable/liburing-dev/io_uring_enter2.2.en.html) | Ring | [Enter2](syscall.go) |  | :heavy_check_mark: |
| [io_uring_for_each_cqe](https://manpages.debian.org/unstable/liburing-dev/io_uring_for_each_cqe.3.en.html) | Ring | [ForEachCQE](lib.go) |  | :heavy_check_mark: |
| [io_uring_free_buf_ring](https://manpages.debian.org/unstable/liburing-dev/io_uring_free_buf_ring.3.en.html) | Ring | [FreeBufRing](setup.ho) |  | :heavy_check_mark: |
| [io_uring_free_probe](https://manpages.debian.org/unstable/liburing-dev/io_uring_free_probe.3.en.html) |  |  | :heavy_exclamation_mark:unnecessary | :heavy_multiplication_x: |
| [io_uring_get_events](https://manpages.debian.org/unstable/liburing-dev/io_uring_get_events.3.en.html) | Ring | [GetEvents](queue.go) |  | :heavy_check_mark: |
| [io_uring_get_probe](https://manpages.debian.org/unstable/liburing-dev/io_uring_get_probe.3.en.html) |  | [GetProbe](probe.go) |  | :heavy_check_mark: |
| io_uring_get_probe_ring | Ring | [GetProbeRing](probe.go) |  | :heavy_check_mark: |
| [io_uring_get_sqe](https://manpages.debian.org/unstable/liburing-dev/io_uring_get_sqe.3.en.html) | Ring | [GetSQE](lib.go) |  | :heavy_check_mark: |
| [io_uring_major_version](https://manpages.debian.org/unstable/liburing-dev/io_uring_major_version.3.en.html) |  | [MajorVersion](version.go) |  | :heavy_check_mark: |
| [io_uring_minor_version](https://manpages.debian.org/unstable/liburing-dev/io_uring_minor_version.3.en.html) |  | [MinorVersion](version.go) |  | :heavy_check_mark: |
| [io_uring_opcode_supported](https://manpages.debian.org/unstable/liburing-dev/io_uring_opcode_supported.3.en.html) | Probe | [IsSupported](probe.go) |  | :heavy_check_mark: |
| [io_uring_peek_cqe](https://manpages.debian.org/unstable/liburing-dev/io_uring_peek_cqe.3.en.html) | Ring | [PeekCQE](lib.go) |  | :heavy_check_mark: |
| [io_uring_prep_accept](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_accept.3.en.html) | SubmissionQueueEntry | [PrepareAccept](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_accept_direct](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_accept_direct.3.en.html) | SubmissionQueueEntry | [PrepareAcceptDirect](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_cancel](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_cancel.3.en.html) | SubmissionQueueEntry | [PrepareCancel](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_cancel64](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_cancel64.3.en.html) | SubmissionQueueEntry | [PrepareCancel64](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_close](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_close.3.en.html) | SubmissionQueueEntry | [PrepareClose](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_close_direct](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_close_direct.3.en.html) | SubmissionQueueEntry | [PrepareCloseDirect](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_connect](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_connect.3.en.html) | SubmissionQueueEntry | [PrepareConnect](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_fadvise](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_fadvise.3.en.html) | SubmissionQueueEntry | [PrepareFadvise](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_fallocate](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_fallocate.3.en.html) | SubmissionQueueEntry | [PrepareFallocate](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_fgetxattr](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_fgetxattr.3.en.html) | SubmissionQueueEntry | [PrepareFgetxattr](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_files_update](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_files_update.3.en.html) | SubmissionQueueEntry | [PrepareFilesUpdate](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_fsetxattr](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_fsetxattr.3.en.html) | SubmissionQueueEntry | [PrepareFsetxattr](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_fsync](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_fsync.3.en.html) | SubmissionQueueEntry | [PrepareFsync](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_getxattr](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_getxattr.3.en.html) | SubmissionQueueEntry | [PrepareGetxattr](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_link](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_link.3.en.html) | SubmissionQueueEntry | [PrepareLink](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_link_timeout](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_link_timeout.3.en.html) | SubmissionQueueEntry | [PrepareLinkTimeout](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_linkat](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_linkat.3.en.html) | SubmissionQueueEntry | [PrepareLinkat](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_madvise](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_madvise.3.en.html) | SubmissionQueueEntry | [PrepareMadvise](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_mkdir](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_mkdir.3.en.html) | SubmissionQueueEntry | [PrepareMkdir](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_mkdirat](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_mkdirat.3.en.html) | SubmissionQueueEntry | [PrepareMkdirat](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_msg_ring](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_msg_ring.3.en.html) | SubmissionQueueEntry | [PrepareMsgRing](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_msg_ring_cqe_flags](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_msg_ring_cqe_flags.3.en.html) | SubmissionQueueEntry | [PrepareMsgRingCqeFlags](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_msg_ring_fd](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_msg_ring_fd.3.en.html) | SubmissionQueueEntry | [PrepareMsgRingFd](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_msg_ring_fd_alloc](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_msg_ring_fd_alloc.3.en.html) | SubmissionQueueEntry | [PrepareMsgRingFdAlloc](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_multishot_accept](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_multishot_accept.3.en.html) | SubmissionQueueEntry | [PrepareMultishotAccept](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_multishot_accept_direct](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_multishot_accept_direct.3.en.html) | SubmissionQueueEntry | [PrepareMultishotAcceptDirect](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_nop](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_nop.3.en.html) | SubmissionQueueEntry | [PrepareNop](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_openat](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_openat.3.en.html) | SubmissionQueueEntry | [PrepareOpenat](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_openat2](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_openat2.3.en.html) | SubmissionQueueEntry | [PrepareOpenat2](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_openat2_direct](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_openat2_direct.3.en.html) | SubmissionQueueEntry | [PrepareOpenat2Direct](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_openat_direct](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_openat_direct.3.en.html) | SubmissionQueueEntry | [PrepareOpenatDirect](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_poll_add](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_poll_add.3.en.html) | SubmissionQueueEntry | [PreparePollAdd](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_poll_multishot](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_poll_multishot.3.en.html) | SubmissionQueueEntry | [PreparePollMultishot](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_poll_remove](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_poll_remove.3.en.html) | SubmissionQueueEntry | [PreparePollRemove](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_poll_update](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_poll_update.3.en.html) | SubmissionQueueEntry | [PreparePollUpdate](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_provide_buffers](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_provide_buffers.3.en.html) | SubmissionQueueEntry | [PrepareProvideBuffers](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_read](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_read.3.en.html) | SubmissionQueueEntry | [PrepareRead](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_read_fixed](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_read_fixed.3.en.html) | SubmissionQueueEntry | [PrepareReadFixed](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_readv](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_readv.3.en.html) | SubmissionQueueEntry | [PrepareReadv](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_readv2](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_readv2.3.en.html) | SubmissionQueueEntry | [PrepareReadv2](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_recv](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_recv.3.en.html) | SubmissionQueueEntry | [PrepareRecv](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_recv_multishot](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_recv_multishot.3.en.html) | SubmissionQueueEntry | [PrepareRecvMultishot](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_recvmsg](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_recvmsg.3.en.html) | SubmissionQueueEntry | [PrepareRecvMsg](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_recvmsg_multishot](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_recvmsg_multishot.3.en.html) | SubmissionQueueEntry | [PrepareRecvMsgMultishot](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_remove_buffers](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_remove_buffers.3.en.html) | SubmissionQueueEntry | [PrepareRemoveBuffers](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_rename](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_rename.3.en.html) | SubmissionQueueEntry | [PrepareRename](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_renameat](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_renameat.3.en.html) | SubmissionQueueEntry | [PrepareRenameat](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_send](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_send.3.en.html) | SubmissionQueueEntry | [PrepareSend](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_send_set_addr](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_send_set_addr.3.en.html) | SubmissionQueueEntry | [PrepareSendSetAddr](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_send_zc](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_send_zc.3.en.html) | SubmissionQueueEntry | [PrepareSendZC](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_send_zc_fixed](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_send_zc_fixed.3.en.html) | SubmissionQueueEntry | [PrepareSendZCFixed](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_sendmsg](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_sendmsg.3.en.html) | SubmissionQueueEntry | [PrepareSendMsg](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_sendmsg_zc](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_sendmsg_zc.3.en.html) | SubmissionQueueEntry | [PrepareSendmsgZC](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_sendto](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_sendto.3.en.html) | SubmissionQueueEntry | [PrepareSendto](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_setxattr](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_setxattr.3.en.html) | SubmissionQueueEntry | [PrepareSetxattr](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_shutdown](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_shutdown.3.en.html) | SubmissionQueueEntry | [PrepareShutdown](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_socket](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_socket.3.en.html) | SubmissionQueueEntry | [PrepareSocket](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_socket_direct](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_socket_direct.3.en.html) | SubmissionQueueEntry | [PrepareSocketDirect](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_socket_direct_alloc](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_socket_direct_alloc.3.en.html) | SubmissionQueueEntry | [PrepareSocketDirectAlloc](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_splice](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_splice.3.en.html) | SubmissionQueueEntry | [PrepareSplice](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_statx](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_statx.3.en.html) | SubmissionQueueEntry | [PrepareStatx](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_symlink](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_symlink.3.en.html) | SubmissionQueueEntry | [PrepareSymlink](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_symlinkat](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_symlinkat.3.en.html) | SubmissionQueueEntry | [PrepareSymlinkat](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_sync_file_range](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_sync_file_range.3.en.html) | SubmissionQueueEntry | [PrepareSyncFileRange](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_tee](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_tee.3.en.html) | SubmissionQueueEntry | [PrepareTee](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_timeout](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_timeout.3.en.html) | SubmissionQueueEntry | [PrepareTimeout](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_timeout_remove](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_timeout_remove.3.en.html) | SubmissionQueueEntry | [PrepareTimeoutRemove](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_timeout_update](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_timeout_update.3.en.html) | SubmissionQueueEntry | [PrepareTimeoutUpdate](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_unlink](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_unlink.3.en.html) | SubmissionQueueEntry | [PrepareUnlink](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_unlinkat](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_unlinkat.3.en.html) | SubmissionQueueEntry | [PrepareUnlinkat](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_write](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_write.3.en.html) | SubmissionQueueEntry | [PrepareWrite](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_write_fixed](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_write_fixed.3.en.html) | SubmissionQueueEntry | [PrepareWriteFixed](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_writev](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_writev.3.en.html) | SubmissionQueueEntry | [PrepareWritev](prepare.go) |  | :heavy_check_mark: |
| [io_uring_prep_writev2](https://manpages.debian.org/unstable/liburing-dev/io_uring_prep_writev2.3.en.html) | SubmissionQueueEntry | [PrepareWritev2](prepare.go) |  | :heavy_check_mark: |
| [io_uring_queue_exit](https://manpages.debian.org/unstable/liburing-dev/io_uring_queue_exit.3.en.html) | Ring | [QueueExit](setup.go) |  | :heavy_check_mark: |
| [io_uring_queue_init](https://manpages.debian.org/unstable/liburing-dev/io_uring_queue_init.3.en.html) | Ring | [QueueInit](setup.go) |  | :heavy_check_mark: |
| [io_uring_queue_init_params](https://manpages.debian.org/unstable/liburing-dev/io_uring_queue_init_params.3.en.html) | Ring | [QueueInitParams](setup.go) |  | :heavy_check_mark: |
| [io_uring_recvmsg_cmsg_firsthdr](https://manpages.debian.org/unstable/liburing-dev/io_uring_recvmsg_cmsg_firsthdr.3.en.html) | RecvmsgOut | [CmsgFirsthdr](recvmsg.go) |  | :heavy_check_mark: |
| [io_uring_recvmsg_cmsg_nexthdr](https://manpages.debian.org/unstable/liburing-dev/io_uring_recvmsg_cmsg_nexthdr.3.en.html) | RecvmsgOut | [CmsgNexthdr](recvmsg.go) |  | :heavy_check_mark: |
| [io_uring_recvmsg_name](https://manpages.debian.org/unstable/liburing-dev/io_uring_recvmsg_name.3.en.html) | RecvmsgOut | [Name](recvmsg.go) |  | :heavy_check_mark: |
| [io_uring_recvmsg_payload](https://manpages.debian.org/unstable/liburing-dev/io_uring_recvmsg_payload.3.en.html) |  RecvmsgOut| [Payload](recvmsg.go) |  | :heavy_check_mark: |
| [io_uring_recvmsg_payload_length](https://manpages.debian.org/unstable/liburing-dev/io_uring_recvmsg_payload_length.3.en.html) | RecvmsgOut | [PayloadLength](recvmsg.go) |  | :heavy_check_mark: |
| [io_uring_recvmsg_validate](https://manpages.debian.org/unstable/liburing-dev/io_uring_recvmsg_validate.3.en.html) | RecvmsgOut | [RecvmsgValidate](recvmsg.go) |  | :heavy_check_mark: |
| [io_uring_register](https://manpages.debian.org/unstable/liburing-dev/io_uring_register.2.en.html) | Ring | [Register](syscall.go) |  | :heavy_check_mark: |
| [io_uring_register_buf_ring](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_buf_ring.3.en.html) | Ring | [RegisterBufferRing](register.go) |  | :heavy_check_mark: |
| [io_uring_register_buffers](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_buffers.3.en.html) | Ring | [RegisterBuffers](register.go) |  | :heavy_check_mark: |
| [io_uring_register_buffers_sparse](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_buffers_sparse.3.en.html) | Ring | [RegisterBuffersSparse](register.go) |  | :heavy_check_mark: |
| [io_uring_register_buffers_tags](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_buffers_tags.3.en.html) | Ring | [RegisterBuffersTags](register.go) |  | :heavy_check_mark: |
| [io_uring_register_buffers_update_tag](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_buffers_update_tag.3.en.html) | Ring | [RegisterBuffersUpdateTag](register.go) |  | :heavy_check_mark: |
| [io_uring_register_eventfd](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_eventfd.3.en.html) | Ring | [RegisterEventFd](register.go) |  | :heavy_check_mark: |
| [io_uring_register_eventfd_async](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_eventfd_async.3.en.html) | Ring | [RegisterEventFdAsync](register.go) |  | :heavy_check_mark: |
| [io_uring_register_file_alloc_range](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_file_alloc_range.3.en.html) | Ring | [RegisterFileAllocRange](register.go) |  | :heavy_check_mark: |
| [io_uring_register_files](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_files.3.en.html) | Ring | [RegisterFiles](register.go) |  | :heavy_check_mark: |
| [io_uring_register_files_sparse](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_files_sparse.3.en.html) | Ring | [RegisterFilesSparse](register.go) |  | :heavy_check_mark: |
| [io_uring_register_files_tags](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_files_tags.3.en.html) | Ring | [RegisterFilesTags](register.go) |  | :heavy_check_mark: |
| [io_uring_register_files_update](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_files_update.3.en.html) | Ring | [RegisterFilesUpdate](register.go) |  | :heavy_check_mark: |
| [io_uring_register_files_update_tag](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_files_update_tag.3.en.html) | Ring | [RegisterFilesUpdateTag](register.go) |  | :heavy_check_mark: |
| [io_uring_register_iowq_aff](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_iowq_aff.3.en.html) | Ring | [RegisterIOWQAff](register.go) |  | :heavy_check_mark: |
| [io_uring_register_iowq_max_workers](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_iowq_max_workers.3.en.html) | Ring | [RegisterIOWQMaxWorkers](register.go) |  | :heavy_check_mark: |
| [io_uring_register_ring_fd](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_ring_fd.3.en.html) | Ring | [RegisterRingFd](register.go) |  | :heavy_check_mark: |
| [io_uring_register_sync_cancel](https://manpages.debian.org/unstable/liburing-dev/io_uring_register_sync_cancel.3.en.html) | Ring | [RegisterSyncCancel](register.go) |  | :heavy_check_mark: |
| [io_uring_setup](https://manpages.debian.org/unstable/liburing-dev/io_uring_setup.2.en.html) |  | [Setup](syscall.go) |  | :heavy_check_mark: |
| [io_uring_setup_buf_ring](https://manpages.debian.org/unstable/liburing-dev/io_uring_setup_buf_ring.3.en.html) | Ring | [SetupBufRing](setup.go) |  | :heavy_check_mark: |
| [io_uring_sq_ready](https://manpages.debian.org/unstable/liburing-dev/io_uring_sq_ready.3.en.html) | Ring | [SQReady](lib.go) |  | :heavy_check_mark: |
| [io_uring_sq_space_left](https://manpages.debian.org/unstable/liburing-dev/io_uring_sq_space_left.3.en.html) | Ring | [SQSpaceLeft](lib.go) |  | :heavy_check_mark: |
| [io_uring_sqe_set_data](https://manpages.debian.org/unstable/liburing-dev/io_uring_sqe_set_data.3.en.html) | SubmissionQueueEntry | [SetData](lib.go) |  | :heavy_check_mark: |
| [io_uring_sqe_set_data64](https://manpages.debian.org/unstable/liburing-dev/io_uring_sqe_set_data64.3.en.html) | SubmissionQueueEntry | [SetData64](lib.go) |  | :heavy_check_mark: |
| [io_uring_sqe_set_flags](https://manpages.debian.org/unstable/liburing-dev/io_uring_sqe_set_flags.3.en.html) | SubmissionQueueEntry | [SetFlags](lib.go) |  | :heavy_check_mark: |
| [io_uring_sqring_wait](https://manpages.debian.org/unstable/liburing-dev/io_uring_sqring_wait.3.en.html) | Ring | [SQRingWait](lib.go) |  | :heavy_check_mark: |
| [io_uring_submit](https://manpages.debian.org/unstable/liburing-dev/io_uring_submit.3.en.html) | Ring | [Submit](queue.go) |  | :heavy_check_mark: |
| [io_uring_submit_and_get_events](https://manpages.debian.org/unstable/liburing-dev/io_uring_submit_and_get_events.3.en.html) | Ring | [SubmitAndGetEvents](queue.go) |  | :heavy_check_mark: |
| [io_uring_submit_and_wait](https://manpages.debian.org/unstable/liburing-dev/io_uring_submit_and_wait.3.en.html) | Ring | [SubmitAndWait](queue.go) |  | :heavy_check_mark: |
| [io_uring_submit_and_wait_timeout](https://manpages.debian.org/unstable/liburing-dev/io_uring_submit_and_wait_timeout.3.en.html) | Ring | [SubmitAndWaitTimeout](queue.go) |  | :heavy_check_mark: |
| [io_uring_unregister_buf_ring](https://manpages.debian.org/unstable/liburing-dev/io_uring_unregister_buf_ring.3.en.html) | Ring | [UnregisterBufferRing](register.go) |  | :heavy_check_mark: |
| [io_uring_unregister_buffers](https://manpages.debian.org/unstable/liburing-dev/io_uring_unregister_buffers.3.en.html) | Ring | [UnregisterBuffers](register.go) |  | :heavy_check_mark: |
| [io_uring_unregister_eventfd](https://manpages.debian.org/unstable/liburing-dev/io_uring_unregister_eventfd.3.en.html) | Ring | [UnregisterEventFd](register.go) |  | :heavy_check_mark: |
| [io_uring_unregister_files](https://manpages.debian.org/unstable/liburing-dev/io_uring_unregister_files.3.en.html) | Ring | [UnregisterFiles](register.go) |  | :heavy_check_mark: |
| [io_uring_unregister_iowq_aff](https://manpages.debian.org/unstable/liburing-dev/io_uring_unregister_iowq_aff.3.en.html) | Ring | [UnregisterIOWQAff](register.go) |  | :heavy_check_mark: |
| [io_uring_unregister_ring_fd](https://manpages.debian.org/unstable/liburing-dev/io_uring_unregister_ring_fd.3.en.html) | Ring | [UnregisterRingFd](register.go) |  | :heavy_check_mark: |
| [io_uring_wait_cqe](https://manpages.debian.org/unstable/liburing-dev/io_uring_wait_cqe.3.en.html) | Ring | [WaitCQE](lib.go) |  | :heavy_check_mark: |
| [io_uring_wait_cqe_nr](https://manpages.debian.org/unstable/liburing-dev/io_uring_wait_cqe_nr.3.en.html) | Ring | [WaitCQENr](lib.go) |  | :heavy_check_mark: |
| [io_uring_wait_cqe_timeout](https://manpages.debian.org/unstable/liburing-dev/io_uring_wait_cqe_timeout.3.en.html) | Ring | [WaitCQETimeout](queue.go) |  | :heavy_check_mark: |
| [io_uring_wait_cqes](https://manpages.debian.org/unstable/liburing-dev/io_uring_wait_cqes.3.en.html) | Ring | [WaitCQEs](queue.go) |  | :heavy_check_mark: |

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## License

Distributed under the MIT License. See `LICENSE` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>


## Contact

Paweł Gaczyński - [LinkedIn](http://linkedin.com/in/pawel-gaczynski)

<p align="right">(<a href="#readme-top">back to top</a>)</p>


## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>