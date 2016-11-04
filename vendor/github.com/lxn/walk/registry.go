// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"syscall"
	"unsafe"
)

import (
	"github.com/lxn/win"
)

type RegistryKey struct {
	hKey win.HKEY
}

func ClassesRootKey() *RegistryKey {
	return &RegistryKey{win.HKEY_CLASSES_ROOT}
}

func CurrentUserKey() *RegistryKey {
	return &RegistryKey{win.HKEY_CURRENT_USER}
}

func LocalMachineKey() *RegistryKey {
	return &RegistryKey{win.HKEY_LOCAL_MACHINE}
}

func RegistryKeyString(rootKey *RegistryKey, subKeyPath, valueName string) (value string, err error) {
	var hKey win.HKEY
	if win.RegOpenKeyEx(
		rootKey.hKey,
		syscall.StringToUTF16Ptr(subKeyPath),
		0,
		win.KEY_READ,
		&hKey) != win.ERROR_SUCCESS {

		return "", newError("RegistryKeyString: Failed to open subkey.")
	}
	defer win.RegCloseKey(hKey)

	var typ uint32
	var data []uint16
	var bufSize uint32

	if win.ERROR_SUCCESS != win.RegQueryValueEx(
		hKey,
		syscall.StringToUTF16Ptr(valueName),
		nil,
		&typ,
		nil,
		&bufSize) {

		return "", newError("RegQueryValueEx #1")
	}

	data = make([]uint16, bufSize/2+1)

	if win.ERROR_SUCCESS != win.RegQueryValueEx(
		hKey,
		syscall.StringToUTF16Ptr(valueName),
		nil,
		&typ,
		(*byte)(unsafe.Pointer(&data[0])),
		&bufSize) {

		return "", newError("RegQueryValueEx #2")
	}

	return syscall.UTF16ToString(data), nil
}

func RegistryKeyUint32(rootKey *RegistryKey, subKeyPath, valueName string) (value uint32, err error) {
	var hKey win.HKEY
	if win.RegOpenKeyEx(
		rootKey.hKey,
		syscall.StringToUTF16Ptr(subKeyPath),
		0,
		win.KEY_READ,
		&hKey) != win.ERROR_SUCCESS {

		return 0, newError("RegistryKeyUint32: Failed to open subkey.")
	}
	defer win.RegCloseKey(hKey)

	bufSize := uint32(4)

	if win.ERROR_SUCCESS != win.RegQueryValueEx(
		hKey,
		syscall.StringToUTF16Ptr(valueName),
		nil,
		nil,
		(*byte)(unsafe.Pointer(&value)),
		&bufSize) {

		return 0, newError("RegQueryValueEx")
	}

	return
}
