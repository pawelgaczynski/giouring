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
	"crypto/rand"
	"os"
	"syscall"
	"testing"
	"time"

	. "github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

const (
	fadviceFileSize = 1024
	fadviceFileName = "fadviceFile"
	fadviceLoops    = 100
	fadviceMinLoops = 10
)

func doRead(t *testing.T, fd int, buf []byte) int64 {
	off, err := syscall.Seek(fd, 0, unix.SEEK_SET)
	NoError(t, err)
	Zero(t, off)

	now := time.Now().UnixMicro()
	ret, err := syscall.Read(fd, buf)
	timeElapsed := time.Now().UnixMicro() - now

	NoError(t, err)
	Equal(t, fadviceFileSize, ret)

	return timeElapsed
}

func doFadvise(t *testing.T, ring *Ring, fd int, offset uint64, length, advice int) bool {
	sqe := ring.GetSQE()
	NotNil(t, sqe)

	sqe.PrepareFadvise(fd, offset, length, advice)
	sqe.UserData = uint64(advice)
	submit, err := ring.SubmitAndWait(1)
	NoError(t, err)
	Equal(t, uint(1), submit)

	cqe, err := ring.WaitCQE()
	NoError(t, err)
	NotNil(t, cqe)

	ret := int(cqe.Res)
	if -ret == int(syscall.EINVAL) || -ret == int(syscall.EBADF) {
		return false
	}

	ring.CQESeen(cqe)

	return true
}

func testFadvice(t *testing.T, ring *Ring, filepath string) (bool, bool) {
	fd, err := syscall.Open(filepath, os.O_RDONLY, 0)
	NoError(t, err)

	buffer := make([]byte, fadviceFileSize)

	cachedRead := doRead(t, fd, buffer)

	supported := doFadvise(t, ring, fd, 0, fadviceFileSize, unix.FADV_DONTNEED)
	if !supported {
		return false, false
	}

	uncachedRead := doRead(t, fd, buffer)
	NoError(t, err)

	doFadvise(t, ring, fd, 0, fadviceFileSize, unix.FADV_DONTNEED)

	doFadvise(t, ring, fd, 0, fadviceFileSize, unix.FADV_WILLNEED)

	_ = syscall.Fsync(fd)

	cachedRead2 := doRead(t, fd, buffer)

	return cachedRead < uncachedRead && cachedRead2 < uncachedRead, true
}

func TestFadvice(t *testing.T) {
	ring, err := CreateRing(16)
	NotNil(t, ring)
	NoError(t, err)

	defer ring.QueueExit()

	file, err := os.Create("fileToFadvice")
	NoError(t, err)

	defer os.Remove(file.Name())

	data := make([]byte, fadviceFileSize)
	_, err = rand.Read(data)
	NoError(t, err)

	n, err := file.Write(data)
	NoError(t, err)
	Equal(t, fadviceFileSize, n)

	var (
		good int
		bad  int
	)
	for i := 0; i < fadviceLoops; i++ {
		res, supported := testFadvice(t, ring, file.Name())
		if !supported {
			break
		}
		if res {
			good++
		} else {
			bad++
		}
		if i >= fadviceMinLoops && bad == 0 {
			break
		}
	}

	Greater(t, good, bad)
}
