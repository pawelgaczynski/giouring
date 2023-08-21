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
	"testing"

	. "github.com/stretchr/testify/require"
)

func TestParseKernelVersion(t *testing.T) {
	// Test case 1: valid kernel version string
	versionString := "5.10.0-rc1"
	expectedVersion := &KernelVersion{Kernel: 5, Major: 10, Minor: 0, Flavor: "-rc1"}
	version, err := parseKernelVersion(versionString)
	NoError(t, err)
	Equal(t, expectedVersion, version)

	// Test case 2: invalid kernel version string
	versionString = "invalid"
	version, err = parseKernelVersion(versionString)
	Error(t, err)
	Nil(t, version)
}

func TestCompareKernelVersion(t *testing.T) {
	// Test case 1: aVersion > bVersion
	aVersion := KernelVersion{Kernel: 5, Major: 10, Minor: 0}
	bVersion := KernelVersion{Kernel: 5, Major: 9, Minor: 0}
	result := CompareKernelVersion(aVersion, bVersion)
	Equal(t, 1, result)

	// Test case 2: aVersion < bVersion
	aVersion = KernelVersion{Kernel: 5, Major: 8, Minor: 0}
	bVersion = KernelVersion{Kernel: 5, Major: 9, Minor: 0}
	result = CompareKernelVersion(aVersion, bVersion)
	Equal(t, -1, result)

	// Test case 3: aVersion == bVersion
	aVersion = KernelVersion{Kernel: 5, Major: 9, Minor: 0}
	bVersion = KernelVersion{Kernel: 5, Major: 9, Minor: 0}
	result = CompareKernelVersion(aVersion, bVersion)
	Equal(t, 0, result)
}
