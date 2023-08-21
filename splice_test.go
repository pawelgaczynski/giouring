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
	"io"
	"os"
	"testing"

	. "github.com/stretchr/testify/require"
)

func TestSplice(t *testing.T) {
	testNewFramework(t, ringInitParams{}, testScenario{
		setup: func(ctx testContext) {
			file1, err := os.CreateTemp("", "splice_1")
			NoError(t, err)
			ctx["file1"] = file1
			err = os.WriteFile(file1.Name(), []byte("test"), 0o600)
			NoError(t, err)

			file2, err := os.CreateTemp("", "splice_1")
			NoError(t, err)
			ctx["file2"] = file2

			pipeR, pipeW, err := os.Pipe()
			NoError(t, err)
			ctx["pipeR"] = pipeR
			ctx["pipeW"] = pipeW
		},

		prepares: []func(*testing.T, testContext, *SubmissionQueueEntry){
			func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
				file1, ok := ctx["file1"].(*os.File)
				True(t, ok)
				pipeW, ok := ctx["pipeW"].(*os.File)
				True(t, ok)
				sqe.PrepareSplice(int(file1.Fd()), -1, int(pipeW.Fd()), -1, 4, 0)
				sqe.UserData = 1
			},
			func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
				file2, ok := ctx["file2"].(*os.File)
				True(t, ok)
				pipeR, ok := ctx["pipeR"].(*os.File)
				True(t, ok)
				sqe.PrepareSplice(int(pipeR.Fd()), -1, int(file2.Fd()), -1, 4, 0)
				sqe.UserData = 2
			},
		},

		result: func(t *testing.T, ctx testContext, cqes []*CompletionQueueEvent) {
			for _, cqe := range cqes {
				Contains(t, []uint64{1, 2}, cqe.UserData)
				Equal(t, cqe.Res, int32(4))
			}
		},

		assert: func(t *testing.T, ctx testContext) {
			val, ok := ctx["file2"]
			True(t, ok)
			file, ok := val.(*os.File)
			True(t, ok)
			data, err := os.ReadFile(file.Name())
			Error(t, io.EOF, err)
			Equal(t, 4, len(data))
		},

		cleanup: func(ctx testContext) {
			files := []string{"file1", "file2"}
			for i := 0; i < len(files); i++ {
				if val, ok := ctx[files[i]]; ok {
					file, fileOk := val.(*os.File)
					True(t, fileOk)
					file.Close()
					os.Remove(file.Name())
				}
			}
		},
	})
}

func TestTee(t *testing.T) {
	testNewFramework(t, ringInitParams{}, testScenario{
		setup: func(ctx testContext) {
			pipe1R, pipe1W, err := os.Pipe()
			NoError(t, err)
			ctx["pipe1R"] = pipe1R
			ctx["pipe1W"] = pipe1W

			pipe2R, pipe2W, err := os.Pipe()
			NoError(t, err)
			ctx["pipe2R"] = pipe2R
			ctx["pipe2W"] = pipe2W

			n, err := pipe1W.WriteString("data")
			NoError(t, err)
			Equal(t, 4, n)
		},

		prepares: []func(*testing.T, testContext, *SubmissionQueueEntry){
			func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
				pipe1R, ok := ctx["pipe1R"].(*os.File)
				True(t, ok)
				pipe2W, ok := ctx["pipe2W"].(*os.File)
				True(t, ok)
				sqe.PrepareTee(int(pipe1R.Fd()), int(pipe2W.Fd()), 4, 0)
				sqe.UserData = 1
			},
		},

		result: func(t *testing.T, ctx testContext, cqes []*CompletionQueueEvent) {
			for _, cqe := range cqes {
				Equal(t, cqe.UserData, uint64(1))
				Equal(t, cqe.Res, int32(4))
			}
		},

		assert: func(t *testing.T, ctx testContext) {
			val, ok := ctx["pipe2R"]
			True(t, ok)
			file, ok := val.(*os.File)
			True(t, ok)

			data := make([]byte, 4)
			n, err := file.Read(data)
			Error(t, io.EOF, err)
			Equal(t, 4, n)
			Equal(t, "data", string(data))
		},

		cleanup: func(ctx testContext) {
			files := []string{"file1", "file2"}
			for i := 0; i < len(files); i++ {
				if val, ok := ctx[files[i]]; ok {
					file, fileOk := val.(*os.File)
					True(t, fileOk)
					file.Close()
					os.Remove(file.Name())
				}
			}
		},
	})
}
