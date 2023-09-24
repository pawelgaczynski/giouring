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
	key1   = "user.val1"
	value1 = "value1"
)

var setupFile = func(t *testing.T, ring *Ring, ctx testContext) {
	file, err := os.Create("xattrfile")

	NoError(t, err)
	_, err = file.WriteString("testdata")
	NoError(t, err)
	ctx["file"] = file
	ctx["fd"] = int(file.Fd())
	ctx["path"] = []byte(file.Name())
}

func testXattr(t *testing.T, prepareSetFunc func(*SubmissionQueueEntry, testContext, []byte, []byte),
	prepareGetFunc func(*SubmissionQueueEntry, testContext, []byte, []byte),
) {
	testCase(t, testScenario{
		loop:  true,
		setup: setupFile,

		prepares: []prepare{
			func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
				keyBytes := []byte(key1)
				valueBytes := []byte(value1)
				ctx["keyBytes"] = keyBytes
				ctx["valueBytes"] = valueBytes

				prepareSetFunc(sqe, ctx, keyBytes, valueBytes)

				sqe.UserData = 1
			},
			func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
				keyBytes := []byte(key1)
				valueBuffer := make([]byte, 255)
				ctx["keyBytes"] = keyBytes
				ctx["valueBuffer"] = valueBuffer

				prepareGetFunc(sqe, ctx, keyBytes, valueBuffer)

				sqe.UserData = 2
			},
		},

		result: func(t *testing.T, ctx testContext, cqes []*CompletionQueueEvent) {
			for _, cqe := range cqes {
				Contains(t, []uint64{1, 2}, cqe.UserData)
				switch cqe.UserData {
				case 1:
					Zero(t, cqe.Res)
				case 2:
					Equal(t, int32(6), cqe.Res)
				}
			}
		},

		assert: func(t *testing.T, ctx testContext) {
			valueBuffer, ok := ctx["valueBuffer"].([]byte)
			True(t, ok)
			Equal(t, value1, string(valueBuffer[:6]))
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

func TestFsetFgetxattr(t *testing.T) {
	testXattr(t, func(sqe *SubmissionQueueEntry, ctx testContext, keyBytes []byte, valueBytes []byte) {
		fd, ok := ctx["fd"].(int)
		True(t, ok)
		sqe.PrepareFsetxattr(fd, keyBytes, valueBytes, 0, uint(len(valueBytes)))
	}, func(sqe *SubmissionQueueEntry, ctx testContext, keyBytes []byte, valueBuffer []byte) {
		fd, ok := ctx["fd"].(int)
		True(t, ok)
		sqe.PrepareFgetxattr(fd, keyBytes, valueBuffer)
	})
}

func TestSetGetxattr(t *testing.T) {
	testXattr(t, func(sqe *SubmissionQueueEntry, ctx testContext, keyBytes []byte, valueBytes []byte) {
		path, ok := ctx["path"].([]byte)
		True(t, ok)
		sqe.PrepareSetxattr(keyBytes, valueBytes, path, 0, uint(len(valueBytes)))
	}, func(sqe *SubmissionQueueEntry, ctx testContext, keyBytes []byte, valueBuffer []byte) {
		path, ok := ctx["path"].([]byte)
		True(t, ok)
		sqe.PrepareGetxattr(keyBytes, valueBuffer, path)
	})
}
