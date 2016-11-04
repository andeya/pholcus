// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"syscall"
)

import (
	"github.com/lxn/win"
)

func knownFolderPath(id win.CSIDL) (string, error) {
	var buf [win.MAX_PATH]uint16

	if !win.SHGetSpecialFolderPath(0, &buf[0], id, false) {
		return "", newError("SHGetSpecialFolderPath failed")
	}

	return syscall.UTF16ToString(buf[0:]), nil
}

func AppDataPath() (string, error) {
	return knownFolderPath(win.CSIDL_APPDATA)
}

func CommonAppDataPath() (string, error) {
	return knownFolderPath(win.CSIDL_COMMON_APPDATA)
}

func LocalAppDataPath() (string, error) {
	return knownFolderPath(win.CSIDL_LOCAL_APPDATA)
}

func DriveNames() ([]string, error) {
	bufLen := win.GetLogicalDriveStrings(0, nil)
	if bufLen == 0 {
		return nil, lastError("GetLogicalDriveStrings")
	}
	buf := make([]uint16, bufLen+1)

	bufLen = win.GetLogicalDriveStrings(bufLen+1, &buf[0])
	if bufLen == 0 {
		return nil, lastError("GetLogicalDriveStrings")
	}

	var names []string

	for i := 0; i < len(buf)-2; {
		name := syscall.UTF16ToString(buf[i:])
		names = append(names, name)
		i += len(name) + 1
	}

	return names, nil
}
