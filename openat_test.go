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

var setupOpenatFile = func(t *testing.T, ring *Ring, ctx testContext) {
	file, err := os.Create("openatfile")
	NoError(t, err)
	_, err = file.WriteString("testdata")
	NoError(t, err)
	err = file.Close()
	NoError(t, err)
	ctx["path"] = []byte(file.Name())
}

var setupRegisteredFiles = func(t *testing.T, ring *Ring, ctx testContext) {
	setupOpenatFile(t, ring, ctx)
	files := []int{-1}
	ret, err := ring.RegisterFiles(files)
	NoError(t, err)
	Equal(t, uint(0), ret)
	ctx["files"] = files
}

var openatCleanup = func(ctx testContext) {
	if path, ok := ctx["path"].([]byte); ok {
		os.Remove(string(path))
	}
}

func testOpenat(t *testing.T,
	setupFunc func(*testing.T, *Ring, testContext),
	prepareOpenatFunc func(*SubmissionQueueEntry, testContext),
	prepareReadFunc func(*SubmissionQueueEntry, testContext, []byte),
	openCQEAssert func(*testing.T, int32),
) {
	testCase(t, testScenario{
		loop:  true,
		setup: setupFunc,

		prepares: []prepare{
			func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
				prepareOpenatFunc(sqe, ctx)

				sqe.UserData = 1
			},
			func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
				valueBuffer := make([]byte, 255)
				ctx["valueBuffer"] = valueBuffer

				prepareReadFunc(sqe, ctx, valueBuffer)

				sqe.UserData = 2
			},
		},

		result: func(t *testing.T, ctx testContext, cqes []*CompletionQueueEvent) {
			for _, cqe := range cqes {
				Contains(t, []uint64{1, 2}, cqe.UserData)
				switch cqe.UserData {
				case 1:
					openCQEAssert(t, cqe.Res)
					ctx["fd"] = int(cqe.Res)
				case 2:
					Equal(t, int32(8), cqe.Res)
				}
			}
		},

		assert: func(t *testing.T, ctx testContext) {
			valueBuffer, ok := ctx["valueBuffer"].([]byte)
			True(t, ok)
			Equal(t, "testdata", string(valueBuffer[:8]))
		},

		cleanup: openatCleanup,
	})
}

func TestOpenat(t *testing.T) {
	testOpenat(t, setupFile, func(sqe *SubmissionQueueEntry, ctx testContext) {
		path, ok := ctx["path"].([]byte)
		True(t, ok)
		sqe.PrepareOpenat(unix.AT_FDCWD, path, unix.O_RDONLY, 0)
	}, func(sqe *SubmissionQueueEntry, ctx testContext, valueBuffer []byte) {
		fd, ok := ctx["fd"].(int)
		True(t, ok)
		sqe.PrepareRead(fd, valueBuffer, uint32(len(valueBuffer)), 0)
	}, func(t *testing.T, res int32) {
		Greater(t, res, int32(0))
	})
}

func TestOpenatDirect(t *testing.T) {
	testOpenat(t, setupRegisteredFiles, func(sqe *SubmissionQueueEntry, ctx testContext) {
		path, ok := ctx["path"].([]byte)
		True(t, ok)
		sqe.PrepareOpenatDirect(unix.AT_FDCWD, path, unix.O_RDONLY, 0, 0)
	}, func(sqe *SubmissionQueueEntry, ctx testContext, valueBuffer []byte) {
		sqe.PrepareRead(0, valueBuffer, uint32(len(valueBuffer)), 0)
		sqe.Flags |= SqeFixedFile
	}, func(t *testing.T, res int32) {
		Equal(t, int32(0), res)
	})
}

func TestOpenat2(t *testing.T) {
	testOpenat(t, setupFile, func(sqe *SubmissionQueueEntry, ctx testContext) {
		path, ok := ctx["path"].([]byte)
		True(t, ok)
		var how unix.OpenHow
		how.Mode = unix.O_RDONLY
		ctx["how"] = &how
		sqe.PrepareOpenat2(unix.AT_FDCWD, path, &how)
	}, func(sqe *SubmissionQueueEntry, ctx testContext, valueBuffer []byte) {
		fd, ok := ctx["fd"].(int)
		True(t, ok)
		sqe.PrepareRead(fd, valueBuffer, uint32(len(valueBuffer)), 0)
	}, func(t *testing.T, res int32) {
		Greater(t, res, int32(0))
	})
}

func TestOpenat2Direct(t *testing.T) {
	testOpenat(t, setupRegisteredFiles, func(sqe *SubmissionQueueEntry, ctx testContext) {
		path, ok := ctx["path"].([]byte)
		True(t, ok)
		var how unix.OpenHow
		how.Mode = unix.O_RDONLY
		ctx["how"] = &how
		sqe.PrepareOpenat2Direct(unix.AT_FDCWD, path, &how, 0)
	}, func(sqe *SubmissionQueueEntry, ctx testContext, valueBuffer []byte) {
		sqe.PrepareRead(0, valueBuffer, uint32(len(valueBuffer)), 0)
		sqe.Flags |= SqeFixedFile
	}, func(t *testing.T, res int32) {
		Equal(t, int32(0), res)
	})
}
