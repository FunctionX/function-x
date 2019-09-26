// Based on ssh/terminal:
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux,!appengine darwin freebsd openbsd netbsd

package term

import (
	"syscall"
	"unsafe"

	"fx/chain/logger"
)

// IsTty returns true if the given file descriptor is a terminal.
func IsTty(fd uintptr) bool {
	var termios logger.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, logger.ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}
