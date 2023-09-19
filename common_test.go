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
	"syscall"
	"testing"

	. "github.com/stretchr/testify/require"
)

var port int32 = 8000

func getTestPort() int {
	return int(atomic.AddInt32(&port, 1))
}

func listenSocket(t *testing.T) (int, int) {
	socketFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	Nil(t, err)
	testPort := getTestPort()

	err = syscall.Bind(socketFd, &syscall.SockaddrInet4{
		Port: testPort,
	})
	Nil(t, err)
	err = syscall.Listen(socketFd, 128)
	Nil(t, err)

	return socketFd, testPort
}

type prepare func(*testing.T, testContext, *SubmissionQueueEntry)

type action func(*testing.T, testContext) int

type setup func(*testing.T, *Ring, testContext)

type cleanup func(testContext)

type result func(*testing.T, testContext, []*CompletionQueueEvent)

type assert func(*testing.T, testContext)

type testScenario struct {
	prepares []prepare
	action   action
	setup    setup
	cleanup  cleanup
	result   result
	assert   assert
	loop     bool
}

type testContext map[string]interface{}

func testCase(t *testing.T, scenario testScenario) {
	ring := NewRing()
	NotNil(t, ring)

	var entries uint32 = 16
	err := ring.QueueInit(entries, 0)
	NoError(t, err)

	defer ring.QueueExit()

	context := make(map[string]interface{})
	if scenario.setup != nil {
		scenario.setup(t, ring, context)
	}

	t.Cleanup(func() {
		if scenario.cleanup != nil {
			scenario.cleanup(context)
		}
	})

	var numberOfLoops int
	var numberOfSQEInLoop int

	if scenario.loop {
		numberOfLoops = len(scenario.prepares)
		numberOfSQEInLoop = 1
	} else {
		numberOfLoops = 1
		numberOfSQEInLoop = len(scenario.prepares)
	}

	prepareIdx := 0

	for i := 0; i < numberOfLoops; i++ {
		for j := 0; j < numberOfSQEInLoop; j++ {
			sqe := ring.GetSQE()
			NotNil(t, sqe)

			scenario.prepares[prepareIdx](t, context, sqe)

			t.Logf(">>> %s # Prepared SQE[%d] = %+v\n", t.Name(), prepareIdx, sqe)

			prepareIdx++
		}

		numberOfCQEWait := numberOfSQEInLoop
		submitted, submitErr := ring.Submit()
		NoError(t, submitErr)
		Equal(t, uint(numberOfSQEInLoop), submitted)

		if scenario.action != nil {
			if cqesExpected := scenario.action(t, context); cqesExpected > 0 {
				numberOfCQEWait = cqesExpected
			}
		}

		_, err = ring.WaitCQENr(uint32(numberOfCQEWait))
		NoError(t, err)

		cqes := make([]*CompletionQueueEvent, numberOfCQEWait)

		numberOfCQEs := ring.PeekBatchCQE(cqes)
		Equal(t, uint32(numberOfCQEWait), numberOfCQEs)

		for idx, cqe := range cqes {
			t.Logf("<<< %s # Received CQE[%d] = %+v\n", t.Name(), idx, cqe)
		}

		scenario.result(t, context, cqes)

		ring.CQAdvance(uint32(submitted))
	}

	if scenario.assert != nil {
		scenario.assert(t, context)
	}
}
