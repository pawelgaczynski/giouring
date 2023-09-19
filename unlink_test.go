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
	"syscall"
	"testing"

	. "github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func testUnlink(t *testing.T, prepareFunc prepare) {
	testCase(t, testScenario{
		setup: func(t *testing.T, ring *Ring, ctx testContext) {
			file, err := os.CreateTemp("", "fileToUnlink")
			NoError(t, err)
			n, err := file.WriteString("testdata")
			NoError(t, err)
			Equal(t, 8, n)

			filePath := file.Name()
			fileInfo, err := os.Stat(filePath)
			NoError(t, err)
			NotNil(t, fileInfo)
			filePathBytes := []byte(filePath)
			ctx["filePathBytes"] = filePathBytes
		},

		prepares: []prepare{
			prepareFunc,
		},

		result: func(t *testing.T, ctx testContext, cqes []*CompletionQueueEvent) {
			cqe := cqes[0]
			Equal(t, uint64(1), cqe.UserData)
			Zero(t, cqe.Res)
		},

		assert: func(t *testing.T, ctx testContext) {
			val, ok := ctx["filePathBytes"]
			True(t, ok)
			filePath, ok := val.([]byte)
			True(t, ok)
			fileInfo, err := os.Stat(string(filePath))
			ErrorIs(t, err, syscall.ENOENT)
			Nil(t, fileInfo)
		},
	})
}

func TestUnlinkat(t *testing.T) {
	testUnlink(t, func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
		filePathBytes, ok := ctx["filePathBytes"].([]byte)
		True(t, ok)

		sqe.PrepareUnlinkat(unix.AT_FDCWD, filePathBytes, 0)
		sqe.UserData = 1
	})
}

func TestUnlink(t *testing.T) {
	testUnlink(t, func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
		filePathBytes, ok := ctx["filePathBytes"].([]byte)
		True(t, ok)

		sqe.PrepareUnlink(filePathBytes, 0)
		sqe.UserData = 1
	})
}
