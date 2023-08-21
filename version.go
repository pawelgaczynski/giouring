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
	LibraryVersionMajor = 2
	LibraryVersionMinor = 5
)

// liburing: io_uring_major_version - https://manpages.debian.org/unstable/liburing-dev/io_uring_major_version.3.en.html
func MajorVersion() int {
	return LibraryVersionMajor
}

// liburing: io_uring_minor_version - https://manpages.debian.org/unstable/liburing-dev/io_uring_minor_version.3.en.html
func MinorVersion() int {
	return LibraryVersionMinor
}

// liburing: io_uring_check_version - https://manpages.debian.org/unstable/liburing-dev/io_uring_check_version.3.en.html
func CheckVersion(major, minor int) bool {
	return major > MajorVersion() ||
		(major == MajorVersion() && minor >= MinorVersion())
}
