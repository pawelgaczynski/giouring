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
	"sync/atomic"
	"testing"

	. "github.com/stretchr/testify/require"
)

var port int32 = 8000

func getTestPort() int {
	return int(atomic.AddInt32(&port, 1))
}

type ringInitParams struct {
	entries uint32
	flags   uint32
}

type testScenario struct {
	prepares []func(*testing.T, testContext, *SubmissionQueueEntry)
	setup    func(testContext)
	cleanup  func(testContext)
	result   func(*testing.T, testContext, []*CompletionQueueEvent)
	assert   func(*testing.T, testContext)
	debug    bool
}

type testContext map[string]interface{}

func testNewFramework(t *testing.T, params ringInitParams, scenario testScenario) {
	ring := NewRing()
	NotNil(t, ring)

	var entries uint32 = 16
	if params.entries != 0 {
		entries = params.entries
	}

	err := ring.QueueInit(entries, params.flags)
	NoError(t, err)

	defer ring.QueueExit()

	context := make(map[string]interface{})
	if scenario.setup != nil {
		scenario.setup(context)
	}

	defer func() {
		if scenario.cleanup != nil {
			scenario.cleanup(context)
		}
	}()

	for i := 0; i < len(scenario.prepares); i++ {
		sqe := ring.GetSQE()
		NotNil(t, sqe)

		scenario.prepares[i](t, context, sqe)

		if scenario.debug {
			// nolint: forbidigo
			fmt.Printf(">>> %s # Prepared SQE[%d] = %+v\n", t.Name(), i, sqe)
		}
	}

	numberOfSQEs := uint32(len(scenario.prepares))

	submitted, err := ring.SubmitAndWait(numberOfSQEs)
	NoError(t, err)
	Equal(t, uint(numberOfSQEs), submitted)

	cqes := make([]*CompletionQueueEvent, numberOfSQEs)

	numberOfCQEs := ring.PeekBatchCQE(cqes)
	Equal(t, numberOfSQEs, numberOfCQEs)

	for i, cqe := range cqes {
		if scenario.debug {
			// nolint: forbidigo
			fmt.Printf("<<< %s # Received CQE[%d] = %+v\n", t.Name(), i, cqe)
		}
	}

	scenario.result(t, context, cqes)

	if scenario.assert != nil {
		scenario.assert(t, context)
	}

	ring.CQAdvance(uint32(submitted))
}
