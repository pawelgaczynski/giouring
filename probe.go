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

const (
	probeOpsSize = 256
	// int(OpLast + 1)
)

// liburing: io_uring_get_probe_ring
func (ring *Ring) GetProbeRing() (*Probe, error) {
	probe := &Probe{}
	_, err := ring.RegisterProbe(probe, probeOpsSize)
	if err != nil {
		return nil, err
	}

	return probe, nil
}

const probeEntries = 2

// liburing: io_uring_get_probe - https://manpages.debian.org/unstable/liburing-dev/io_uring_get_probe.3.en.html
func GetProbe() (*Probe, error) {
	ring, err := CreateRing(probeEntries)
	if err != nil {
		return nil, err
	}

	probe, err := ring.GetProbeRing()
	if err != nil {
		return nil, err
	}
	ring.QueueExit()

	return probe, nil
}

// liburing: io_uring_opcode_supported - https://manpages.debian.org/unstable/liburing-dev/io_uring_opcode_supported.3.en.html
func (p Probe) IsSupported(op uint8) bool {
	for i := uint8(0); i < p.OpsLen; i++ {
		if p.Ops[i].Op != op {
			continue
		}

		return p.Ops[i].Flags&opSupported != 0
	}

	return false
}
