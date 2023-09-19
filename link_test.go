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
	"golang.org/x/sys/unix"
)

func testLink(t *testing.T, prepareFunc prepare) {
	testCase(t, testScenario{
		setup: func(t *testing.T, ring *Ring, ctx testContext) {
			file, err := os.Create("fileToLink")
			NoError(t, err)
			n, err := file.WriteString("testdata")
			NoError(t, err)
			Equal(t, 8, n)
			ctx["file"] = file
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
			val, ok := ctx["linkedFilePathBytes"]
			True(t, ok)
			file, ok := val.([]byte)
			True(t, ok)
			data, err := os.ReadFile(string(file))
			Error(t, io.EOF, err)
			Equal(t, 8, len(data))
		},

		cleanup: func(ctx testContext) {
			files := []string{"filePathBytes", "linkedFilePathBytes"}
			for i := 0; i < len(files); i++ {
				if val, ok := ctx[files[i]]; ok {
					file, fileOk := val.([]byte)
					True(t, fileOk)
					err := os.Remove(string(file))
					NoError(t, err)
				}
			}
		},
	})
}

func getLinkeFilePaths(t *testing.T, ctx testContext) ([]byte, []byte) {
	file, ok := ctx["file"].(*os.File)
	True(t, ok)

	linkedFilePath := "fileLinked"
	filePath := file.Name()
	filePathBytes := []byte(filePath)
	linkedFilePathBytes := []byte(linkedFilePath)
	ctx["filePathBytes"] = filePathBytes
	ctx["linkedFilePathBytes"] = linkedFilePathBytes

	return filePathBytes, linkedFilePathBytes
}

func TestLinkat(t *testing.T) {
	testLink(t, func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
		filePathBytes, linkedFilePathBytes := getLinkeFilePaths(t, ctx)

		sqe.PrepareLinkat(unix.AT_FDCWD, filePathBytes, unix.AT_FDCWD, linkedFilePathBytes, 0)
		sqe.UserData = 1
	})
}

func TestLink(t *testing.T) {
	testLink(t, func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
		filePathBytes, linkedFilePathBytes := getLinkeFilePaths(t, ctx)

		sqe.PrepareLink(filePathBytes, linkedFilePathBytes, 0)
		sqe.UserData = 1
	})
}

func TestSymlink(t *testing.T) {
	testLink(t, func(t *testing.T, ctx testContext, sqe *SubmissionQueueEntry) {
		filePathBytes, linkedFilePathBytes := getLinkeFilePaths(t, ctx)

		sqe.PrepareSymlink(filePathBytes, linkedFilePathBytes)
		sqe.UserData = 1
	})
}
