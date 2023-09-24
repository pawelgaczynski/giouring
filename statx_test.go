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

func TestStatx(t *testing.T) {
	testCase(t, testScenario{
		setup: func(t *testing.T, ring *Ring, ctx testContext) {
			file, err := os.CreateTemp("./", "statxfile")
			NoError(t, err)
			_, err = file.WriteString("testdata")
			NoError(t, err)
			file.Close()
			ctx["filepath"] = []byte(file.Name())

			var statx unix.Statx_t
			ctx["statx"] = &statx
		},

		prepares: []prepare{
			func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
				path, ok := ctx["filepath"].([]byte)
				True(t, ok)
				statx, ok := ctx["statx"].(*unix.Statx_t)
				True(t, ok)

				sqe.PrepareStatx(unix.AT_FDCWD, path, 0, unix.STATX_ALL, statx)
				sqe.UserData = 1
			},
		},

		result: func(t *testing.T, ctx testContext, cqes []*CompletionQueueEvent) {
			Equal(t, cqes[0].UserData, uint64(1))
			Zero(t, cqes[0].Res)
		},

		assert: func(t *testing.T, ctx testContext) {
			path, ok := ctx["filepath"].([]byte)
			True(t, ok)
			statx, ok := ctx["statx"].(*unix.Statx_t)
			True(t, ok)

			var osStatx unix.Statx_t
			err := unix.Statx(unix.AT_FDCWD, string(path), 0, unix.STATX_ALL, &osStatx)
			NoError(t, err)

			Equal(t, osStatx, *statx)
		},

		cleanup: func(ctx testContext) {
			if val, ok := ctx["filepath"]; ok {
				path, pathOk := val.([]byte)
				True(t, pathOk)
				os.Remove(string(path))
			}
		},
	})
}
