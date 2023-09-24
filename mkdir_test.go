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
	"golang.org/x/sys/unix"
)

func testMkdir(t *testing.T, prepareFunc prepare) {
	testCase(t, testScenario{
		setup: func(t *testing.T, ring *Ring, ctx testContext) {
			ctx["mkdirPath"] = []byte("testDir")
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
			val, ok := ctx["mkdirPath"]
			True(t, ok)
			mkdirPath, ok := val.([]byte)
			True(t, ok)

			dirInfo, err := os.Stat(string(mkdirPath))
			NoError(t, err)
			NotNil(t, dirInfo)
			True(t, dirInfo.Mode().IsDir())
			Equal(t, os.FileMode(0700), dirInfo.Mode().Perm())
		},

		cleanup: func(ctx testContext) {
			if val, ok := ctx["mkdirPath"]; ok {
				mkdirPath, mkdirPathOk := val.([]byte)
				True(t, mkdirPathOk)
				os.Remove(string(mkdirPath))
			}
		},
	})
}

func TestMkdirkat(t *testing.T) {
	testMkdir(t, func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
		mkdirPath, ok := ctx["mkdirPath"].([]byte)
		True(t, ok)
		sqe.PrepareMkdirat(unix.AT_FDCWD, mkdirPath, 0700)
		sqe.UserData = 1
	})
}

func TestMkdir(t *testing.T) {
	testMkdir(t, func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
		mkdirPath, ok := ctx["mkdirPath"].([]byte)
		True(t, ok)
		sqe.PrepareMkdir(mkdirPath, 0700)
		sqe.UserData = 1
	})
}
