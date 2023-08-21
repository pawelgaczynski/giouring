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

package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/pawelgaczynski/giouring"
)

const (
	ringSize  = 9192
	batchSize = 4096
	benchTime = 10
)

func main() {
	runtime.GOMAXPROCS(1)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ring, ringErr := giouring.CreateRing(ringSize)
	if ringErr != nil {
		log.Panic(ringErr)
	}
	defer ring.QueueExit()

	cqeBuff := make([]*giouring.CompletionQueueEvent, batchSize)

	benchTimeExpected := (time.Second * benchTime).Nanoseconds()
	startTime := time.Now().UnixNano()
	var count uint64

	for {
		for i := 0; i < batchSize; i++ {
			entry := ring.GetSQE()
			if entry == nil {
				log.Panic()
			}

			entry.PrepareNop()
		}
		submitted, err := ring.SubmitAndWait(batchSize)
		if err != nil {
			log.Panic(err)
		}

		if batchSize != int(submitted) {
			log.Panicf("Submitted %d, expected %d", submitted, batchSize)
		}
		peeked := ring.PeekBatchCQE(cqeBuff)
		if batchSize != int(peeked) {
			log.Panicf("Peeked %d, expected %d", peeked, batchSize)
		}

		count += uint64(peeked)

		ring.CQAdvance(uint32(submitted))

		nowTime := time.Now().UnixNano()
		elapsedTime := nowTime - startTime

		if elapsedTime > benchTimeExpected {
			duration := time.Duration(elapsedTime * int64(time.Nanosecond))
			// nolint: forbidigo
			fmt.Println("Submitted ", count, " entries in ", duration, " seconds")
			// nolint: forbidigo
			fmt.Println(count/uint64(duration.Seconds()), " ops/s")

			return
		}
	}
}
