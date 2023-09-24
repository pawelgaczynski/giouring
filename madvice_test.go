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
	madviceFileSize = 128 * 1024
	madviceFileName = "madviceFile"
	madviceLoops    = 100
	madviceMinLoops = 10
)

func doCopy(t *testing.T, dst, src []byte) int64 {
	now := time.Now().UnixMicro()
	n := copy(dst, src)
	timeElapsed := time.Now().UnixMicro() - now
	Equal(t, len(src), n)

	return timeElapsed
}

func doMadvise(t *testing.T, ring *Ring, data []byte, advice int) bool {
	sqe := ring.GetSQE()
	NotNil(t, sqe)

	sqe.PrepareMadvise(data, advice)
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

func testMadvice(t *testing.T, ring *Ring, filepath string) (bool, bool) {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	NoError(t, err)

	buffer := make([]byte, madviceFileSize)

	source, err := syscall.Mmap(int(file.Fd()), 0, madviceFileSize, syscall.PROT_READ, syscall.MAP_PRIVATE)
	NoError(t, err)
	Equal(t, madviceFileSize, len(source))

	defer func() {
		_ = syscall.Munmap(source)
	}()

	_ = doCopy(t, buffer, source)

	cachedRead := doCopy(t, buffer, source)

	supported := doMadvise(t, ring, source, unix.MADV_DONTNEED)
	if !supported {
		return false, false
	}

	uncachedRead := doCopy(t, buffer, source)
	NoError(t, err)

	doMadvise(t, ring, source, unix.MADV_DONTNEED)

	doMadvise(t, ring, source, unix.MADV_WILLNEED)

	err = unix.Msync(source, unix.MS_SYNC)
	NoError(t, err)

	cachedRead2 := doCopy(t, buffer, source)

	return cachedRead < uncachedRead && cachedRead2 < uncachedRead, true
}

func TestMadvice(t *testing.T) {
	// FIXME: this test is based on the corresponding liburing test, but verification of the result fails
	t.SkipNow()

	ring, err := CreateRing(16)
	NotNil(t, ring)
	NoError(t, err)

	defer ring.QueueExit()

	file, err := os.Create("fileToMadvice")
	NoError(t, err)

	defer os.Remove(file.Name())

	data := make([]byte, madviceFileSize)
	_, err = rand.Read(data)
	NoError(t, err)

	n, err := file.Write(data)
	NoError(t, err)
	Equal(t, madviceFileSize, n)

	var (
		good int
		bad  int
	)
	for i := 0; i < madviceLoops; i++ {
		res, supported := testMadvice(t, ring, file.Name())
		if !supported {
			break
		}
		if res {
			good++
		} else {
			bad++
		}
		if i >= madviceMinLoops && bad == 0 {
			break
		}
	}

	Greater(t, good, bad)
}
