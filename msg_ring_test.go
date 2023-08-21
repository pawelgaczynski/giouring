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
	"syscall"
	"testing"

	. "github.com/stretchr/testify/require"
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
