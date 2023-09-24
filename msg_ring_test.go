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
	"os"
	"runtime"
	"syscall"
	"testing"

	. "github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestMsgRingItself(t *testing.T) {
	ring, err := CreateRing(16)
	Nil(t, err)

	defer ring.QueueExit()

	entry := ring.GetSQE()
	NotNil(t, entry)
	entry.PrepareMsgRing(ring.ringFd, 100, 200, 0)
	entry.UserData = 123

	entry = ring.GetSQE()
	NotNil(t, entry)
	entry.PrepareMsgRing(ring.ringFd, 300, 400, 0)
	entry.UserData = 234

	entry = ring.GetSQE()
	NotNil(t, entry)
	entry.PrepareMsgRing(ring.ringFd, 500, 600, 0)
	entry.UserData = 345

	numberOfCQEsSubmitted, err := ring.Submit()
	Nil(t, err)

	if numberOfCQEsSubmitted == 1 {
		cqe, cqeErr := ring.WaitCQE()
		Nil(t, cqeErr)

		if cqe.Res == -int32(syscall.EINVAL) || cqe.Res == -int32(syscall.EOPNOTSUPP) {
			//nolint
			fmt.Println("Skipping test because of no msg support")

			return
		}
	}

	Equal(t, uint(3), numberOfCQEsSubmitted)

	cqes := make([]*CompletionQueueEvent, 128)

	numberOfCQEs := ring.PeekBatchCQE(cqes)
	Equal(t, uint32(6), numberOfCQEs)

	cqe := cqes[0]
	Equal(t, uint64(200), cqe.UserData)
	Equal(t, int32(100), cqe.Res)

	cqe = cqes[1]
	Equal(t, uint64(400), cqe.UserData)
	Equal(t, int32(300), cqe.Res)

	cqe = cqes[2]
	Equal(t, uint64(600), cqe.UserData)
	Equal(t, int32(500), cqe.Res)

	cqe = cqes[3]
	Equal(t, uint64(123), cqe.UserData)
	Equal(t, int32(0), cqe.Res)

	cqe = cqes[4]
	Equal(t, uint64(234), cqe.UserData)
	Equal(t, int32(0), cqe.Res)

	cqe = cqes[5]
	Equal(t, uint64(345), cqe.UserData)
	Equal(t, int32(0), cqe.Res)

	ring.CQAdvance(numberOfCQEs)
}

func TestMsgRing(t *testing.T) {
	senderRing, err := CreateRing(16)
	Nil(t, err)

	defer senderRing.QueueExit()

	receiverRing, err := CreateRing(16)
	Nil(t, err)

	defer receiverRing.QueueExit()

	entry := senderRing.GetSQE()
	NotNil(t, entry)
	entry.PrepareMsgRing(receiverRing.ringFd, 100, 200, 0)

	entry = senderRing.GetSQE()
	NotNil(t, entry)
	entry.PrepareMsgRing(receiverRing.ringFd, 300, 400, 0)

	entry = senderRing.GetSQE()
	NotNil(t, entry)
	entry.PrepareMsgRing(receiverRing.ringFd, 500, 600, 0)

	cqeNr, err := senderRing.Submit()
	Nil(t, err)

	if cqeNr == 1 {
		cqe, cqeErr := senderRing.WaitCQE()
		Nil(t, cqeErr)

		if cqe.Res == -int32(syscall.EINVAL) || cqe.Res == -int32(syscall.EOPNOTSUPP) {
			//nolint
			fmt.Println("Skipping test because of no msg support")

			return
		}
	}

	Equal(t, uint(3), cqeNr)

	cqes := make([]*CompletionQueueEvent, 128)

	numberOfCQEs := receiverRing.PeekBatchCQE(cqes)
	Equal(t, uint32(3), numberOfCQEs)

	cqe := cqes[0]
	Equal(t, uint64(200), cqe.UserData)
	Equal(t, int32(100), cqe.Res)

	cqe = cqes[1]
	Equal(t, uint64(400), cqe.UserData)
	Equal(t, int32(300), cqe.Res)

	cqe = cqes[2]
	Equal(t, uint64(600), cqe.UserData)
	Equal(t, int32(500), cqe.Res)
	receiverRing.CQAdvance(numberOfCQEs)
}

func TestMsgRingCQEFlags(t *testing.T) {
	senderRing, err := CreateRing(16)
	Nil(t, err)

	defer senderRing.QueueExit()

	receiverRing, err := CreateRing(16)
	Nil(t, err)

	defer receiverRing.QueueExit()

	entry := senderRing.GetSQE()
	NotNil(t, entry)
	entry.PrepareMsgRingCqeFlags(receiverRing.ringFd, 100, 200, 0, 1)

	entry = senderRing.GetSQE()
	NotNil(t, entry)
	entry.PrepareMsgRingCqeFlags(receiverRing.ringFd, 300, 400, 0, 1)

	entry = senderRing.GetSQE()
	NotNil(t, entry)
	entry.PrepareMsgRingCqeFlags(receiverRing.ringFd, 500, 600, 0, 1)

	cqeNr, err := senderRing.Submit()
	Nil(t, err)

	if cqeNr == 1 {
		cqe, cqeErr := senderRing.WaitCQE()
		Nil(t, cqeErr)

		if cqe.Res == -int32(syscall.EINVAL) || cqe.Res == -int32(syscall.EOPNOTSUPP) {
			//nolint
			fmt.Println("Skipping test because of no msg support")

			return
		}
	}

	Equal(t, uint(3), cqeNr)

	cqes := make([]*CompletionQueueEvent, 128)

	numberOfCQEs := receiverRing.PeekBatchCQE(cqes)
	Equal(t, uint32(3), numberOfCQEs)

	cqe := cqes[0]
	Equal(t, uint64(200), cqe.UserData)
	Equal(t, int32(100), cqe.Res)
	Equal(t, uint32(100), cqe.Flags)

	cqe = cqes[1]
	Equal(t, uint64(400), cqe.UserData)
	Equal(t, int32(300), cqe.Res)
	Equal(t, uint32(100), cqe.Flags)

	cqe = cqes[2]
	Equal(t, uint64(600), cqe.UserData)
	Equal(t, int32(500), cqe.Res)
	Equal(t, uint32(100), cqe.Flags)
	receiverRing.CQAdvance(numberOfCQEs)
}

var (
	filesize      = 128
	pat      byte = 0x9a
)

func verifyFixedRead(t *testing.T, ring *Ring, fixedFd int, fail bool) {
	buf := make([]byte, filesize)
	entry := ring.GetSQE()
	entry.PrepareRead(fixedFd, buf, uint32(len(buf)), 0)
	entry.Flags |= SqeFixedFile
	submitted, err := ring.Submit()
	NoError(t, err)
	Equal(t, uint(1), submitted)

	cqe, err := ring.WaitCQE()
	NoError(t, err)
	NotNil(t, cqe)
	if fail {
		Equal(t, int32(syscall.EBADF), -cqe.Res)

		return
	}
	Equal(t, int32(filesize), cqe.Res)
	ring.CQESeen(cqe)

	for i := 0; i < filesize; i++ {
		Equal(t, buf[i], pat)
	}
	runtime.KeepAlive(buf)
}

func testMsgRingFd(
	t *testing.T,
	prepareSendMsgFunc func(*SubmissionQueueEntry, int, int, int, uint64),
	targetFdProvider func(int32, int) int,
) {
	file, err := os.Create("msgringfd")
	NoError(t, err)

	defer func() {
		_ = os.Remove(file.Name())
	}()

	buf := make([]byte, filesize)
	for i := range buf {
		buf[i] = pat
	}
	n, err := file.Write(buf)
	Equal(t, n, filesize)
	NoError(t, err)

	err = file.Close()
	NoError(t, err)

	senderRing, err := CreateRing(16)
	Nil(t, err)
	ret, err := senderRing.RegisterFilesSparse(8)
	Nil(t, err)
	Equal(t, uint(0), ret)

	defer senderRing.QueueExit()

	receiverRing, err := CreateRing(16)
	Nil(t, err)
	ret, err = receiverRing.RegisterFilesSparse(8)
	Nil(t, err)
	Equal(t, uint(0), ret)

	defer receiverRing.QueueExit()

	entry := senderRing.GetSQE()
	NotNil(t, entry)
	filepath := []byte(file.Name())
	sourceFd := 0

	entry.PrepareOpenatDirect(unix.AT_FDCWD, filepath, 0, 644, uint32(sourceFd))
	entry.UserData = 1

	submitted, err := senderRing.Submit()
	Nil(t, err)
	Equal(t, uint(1), submitted)
	runtime.KeepAlive(filepath)

	cqe, cqeErr := senderRing.WaitCQE()
	Nil(t, cqeErr)
	Equal(t, int32(0), cqe.Res)
	Equal(t, uint64(1), cqe.UserData)
	Equal(t, uint32(0), cqe.Flags)
	senderRing.CQESeen(cqe)

	verifyFixedRead(t, senderRing, sourceFd, false)

	targetFd := 1
	entry = senderRing.GetSQE()
	prepareSendMsgFunc(entry, receiverRing.ringFd, sourceFd, targetFd, 100)
	entry.UserData = 2

	submitted, err = senderRing.Submit()
	Nil(t, err)
	Equal(t, uint(1), submitted)

	cqe, cqeErr = senderRing.WaitCQE()
	Nil(t, cqeErr)
	Equal(t, int32(0), cqe.Res)
	Equal(t, uint64(2), cqe.UserData)
	Equal(t, uint32(0), cqe.Flags)
	senderRing.CQESeen(cqe)

	cqe, cqeErr = receiverRing.WaitCQE()
	Nil(t, cqeErr)
	Equal(t, int32(0), cqe.Res)
	Equal(t, uint64(100), cqe.UserData)
	Equal(t, uint32(0), cqe.Flags)
	receiverRing.CQESeen(cqe)

	targetFd = targetFdProvider(cqe.Res, targetFd)

	verifyFixedRead(t, receiverRing, targetFd, false)

	entry = senderRing.GetSQE()
	entry.PrepareCloseDirect(uint32(sourceFd))
	entry.UserData = 3

	submitted, err = senderRing.Submit()
	Nil(t, err)
	Equal(t, uint(1), submitted)

	cqe, cqeErr = senderRing.WaitCQE()
	Nil(t, cqeErr)
	Equal(t, int32(0), cqe.Res)
	Equal(t, uint64(3), cqe.UserData)
	senderRing.CQESeen(cqe)

	verifyFixedRead(t, senderRing, sourceFd, true)

	verifyFixedRead(t, receiverRing, targetFd, false)
}

func TestMsgRingFd(t *testing.T) {
	testMsgRingFd(t, func(entry *SubmissionQueueEntry, ringFd, sourceFd, targetFd int, userData uint64) {
		entry.PrepareMsgRingFd(ringFd, sourceFd, targetFd, userData, 0)
	}, func(res int32, fd int) int {
		return fd
	})
}

func TestMsgRingFdAlloc(t *testing.T) {
	testMsgRingFd(t, func(entry *SubmissionQueueEntry, ringFd, sourceFd, targetFd int, userData uint64) {
		entry.PrepareMsgRingFdAlloc(ringFd, sourceFd, userData, 0)
	}, func(res int32, fd int) int {
		return int(res)
	})
}
