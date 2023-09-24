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
	"os"
	"testing"

	. "github.com/stretchr/testify/require"
)

const (
	fallocateSize = 128 * 1024
)

func TestFallocate(t *testing.T) {
	testCase(t, testScenario{
		setup: func(t *testing.T, ring *Ring, ctx testContext) {
			file, err := os.CreateTemp("", "fallocate")
			NoError(t, err)
			_, err = file.WriteString("testdata")
			NoError(t, err)
			ctx["file"] = file
		},

		prepares: []prepare{
			func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
				file, ok := ctx["file"].(*os.File)
				True(t, ok)
				sqe.PrepareFallocate(int(file.Fd()), 0, 0, fallocateSize)
				sqe.UserData = 1
			},
		},

		result: func(t *testing.T, ctx testContext, cqes []*CompletionQueueEvent) {
			Equal(t, cqes[0].UserData, uint64(1))
			Zero(t, cqes[0].Res)
		},

		assert: func(t *testing.T, ctx testContext) {
			file, ok := ctx["file"].(*os.File)
			True(t, ok)
			stat, err := os.Stat(file.Name())
			NoError(t, err)
			Equal(t, fallocateSize, int(stat.Size()))
		},

		cleanup: func(ctx testContext) {
			if val, ok := ctx["file"]; ok {
				file, fileOk := val.(*os.File)
				True(t, fileOk)
				file.Close()
				os.Remove(file.Name())
			}
		},
	})
}
