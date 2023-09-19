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

func testRename(t *testing.T, prepareFunc prepare) {
	testCase(t, testScenario{
		setup: func(t *testing.T, ring *Ring, ctx testContext) {
			file, err := os.CreateTemp("./", "renamefrom")
			NoError(t, err)
			_, err = file.WriteString("testdata")
			NoError(t, err)
			file.Close()
			ctx["oldpath"] = []byte(file.Name())

			newFile, err := os.CreateTemp("./", "renameto")
			NoError(t, err)
			newFile.Close()
			ctx["newpath"] = []byte(newFile.Name())
		},

		prepares: []prepare{
			prepareFunc,
		},

		result: func(t *testing.T, ctx testContext, cqes []*CompletionQueueEvent) {
			Equal(t, cqes[0].UserData, uint64(1))
			Zero(t, cqes[0].Res)
		},

		assert: func(t *testing.T, ctx testContext) {
			newPath, ok := ctx["newpath"].([]byte)
			True(t, ok)
			oldPath, ok := ctx["oldpath"].([]byte)
			True(t, ok)

			_, err := os.Stat(string(oldPath))
			True(t, os.IsNotExist(err))

			data, err := os.ReadFile(string(newPath))
			NoError(t, err)
			Equal(t, []byte("testdata"), data)
		},

		cleanup: func(ctx testContext) {
			files := []string{"oldpath", "newpath"}
			for i := 0; i < len(files); i++ {
				if val, ok := ctx[files[i]]; ok {
					path, pathOk := val.([]byte)
					True(t, pathOk)
					os.Remove(string(path))
				}
			}
		},
	})
}

func TestRename(t *testing.T) {
	testRename(t, func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
		oldpath, ok := ctx["oldpath"].([]byte)
		True(t, ok)
		newpath, ok := ctx["newpath"].([]byte)
		True(t, ok)
		sqe.PrepareRename(oldpath, newpath, 0)
		sqe.UserData = 1
	})
}

func TestRenameat(t *testing.T) {
	testRename(t, func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
		oldpath, ok := ctx["oldpath"].([]byte)
		True(t, ok)
		newpath, ok := ctx["newpath"].([]byte)
		True(t, ok)
		sqe.PrepareRenameat(unix.AT_FDCWD, oldpath, unix.AT_FDCWD, newpath, 0)
		sqe.UserData = 1
	})
}
