// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

import (
	"github.com/lxn/win"
)

type FileDialog struct {
	Title          string
	FilePath       string
	FilePaths      []string
	InitialDirPath string
	Filter         string
	FilterIndex    int
}

func (dlg *FileDialog) show(owner Form, fun func(ofn *win.OPENFILENAME) bool, flags uint32) (accepted bool, err error) {
	ofn := new(win.OPENFILENAME)

	ofn.LStructSize = uint32(unsafe.Sizeof(*ofn))
	if owner != nil {
		ofn.HwndOwner = owner.Handle()
	}

	filter := make([]uint16, len(dlg.Filter)+2)
	copy(filter, syscall.StringToUTF16(dlg.Filter))
	// Replace '|' with the expected '\0'.
	for i, c := range filter {
		if byte(c) == '|' {
			filter[i] = uint16(0)
		}
	}
	ofn.LpstrFilter = &filter[0]
	ofn.NFilterIndex = uint32(dlg.FilterIndex)

	ofn.LpstrInitialDir = syscall.StringToUTF16Ptr(dlg.InitialDirPath)
	ofn.LpstrTitle = syscall.StringToUTF16Ptr(dlg.Title)
	ofn.Flags = win.OFN_FILEMUSTEXIST | flags

	var fileBuf []uint16
	if flags&win.OFN_ALLOWMULTISELECT > 0 {
		fileBuf = make([]uint16, 65536)
	} else {
		fileBuf = make([]uint16, 1024)
		copy(fileBuf, syscall.StringToUTF16(dlg.FilePath))
	}
	ofn.LpstrFile = &fileBuf[0]
	ofn.NMaxFile = uint32(len(fileBuf))

	if !fun(ofn) {
		errno := win.CommDlgExtendedError()
		if errno != 0 {
			err = newError(fmt.Sprintf("Error %d", errno))
		}
		return
	}

	if flags&win.OFN_ALLOWMULTISELECT > 0 {
		split := func() [][]uint16 {
			var parts [][]uint16

			from := 0
			for i, c := range fileBuf {
				if c == 0 {
					if i == from {
						return parts
					}

					parts = append(parts, fileBuf[from:i])
					from = i + 1
				}
			}

			return parts
		}

		parts := split()

		if len(parts) == 1 {
			dlg.FilePaths = []string{syscall.UTF16ToString(parts[0])}
		} else {
			dirPath := syscall.UTF16ToString(parts[0])
			dlg.FilePaths = make([]string, len(parts)-1)

			for i, fp := range parts[1:] {
				dlg.FilePaths[i] = filepath.Join(dirPath, syscall.UTF16ToString(fp))
			}
		}
	} else {
		dlg.FilePath = syscall.UTF16ToString(fileBuf)
	}

	accepted = true

	return
}

func (dlg *FileDialog) ShowOpen(owner Form) (accepted bool, err error) {
	return dlg.show(owner, win.GetOpenFileName, 0)
}

func (dlg *FileDialog) ShowOpenMultiple(owner Form) (accepted bool, err error) {
	return dlg.show(owner, win.GetOpenFileName, win.OFN_ALLOWMULTISELECT|win.OFN_EXPLORER)
}

func (dlg *FileDialog) ShowSave(owner Form) (accepted bool, err error) {
	return dlg.show(owner, win.GetSaveFileName, 0)
}

func (dlg *FileDialog) ShowBrowseFolder(owner Form) (accepted bool, err error) {
	// Calling OleInitialize (or similar) is required for BIF_NEWDIALOGSTYLE.
	if hr := win.OleInitialize(); hr != win.S_OK && hr != win.S_FALSE {
		return false, newError(fmt.Sprint("OleInitialize Error: ", hr))
	}
	defer win.OleUninitialize()

	pathFromPIDL := func(pidl uintptr) (string, error) {
		var path [win.MAX_PATH]uint16
		if !win.SHGetPathFromIDList(pidl, &path[0]) {
			return "", newError("SHGetPathFromIDList failed")
		}

		return syscall.UTF16ToString(path[:]), nil
	}

	// We use this callback to disable the OK button in case of "invalid"
	// selections.
	callback := func(hwnd win.HWND, msg uint32, lp, wp uintptr) uintptr {
		const BFFM_SELCHANGED = 2
		if msg == BFFM_SELCHANGED {
			_, err := pathFromPIDL(lp)
			var enabled uintptr
			if err == nil {
				enabled = 1
			}

			const BFFM_ENABLEOK = win.WM_USER + 101

			win.SendMessage(hwnd, BFFM_ENABLEOK, 0, enabled)
		}

		return 0
	}

	var ownerHwnd win.HWND
	if owner != nil {
		ownerHwnd = owner.Handle()
	}

	// We need to put the initial path into a buffer of at least MAX_LENGTH
	// length, or we may get random crashes.
	var buf [win.MAX_PATH]uint16
	copy(buf[:], syscall.StringToUTF16(dlg.InitialDirPath))

	const BIF_NEWDIALOGSTYLE = 0x00000040

	bi := win.BROWSEINFO{
		HwndOwner: ownerHwnd,
		LpszTitle: syscall.StringToUTF16Ptr(dlg.Title),
		UlFlags:   BIF_NEWDIALOGSTYLE,
		Lpfn:      syscall.NewCallback(callback),
	}

	win.SHParseDisplayName(&buf[0], 0, &bi.PidlRoot, 0, nil)

	pidl := win.SHBrowseForFolder(&bi)
	if pidl == 0 {
		return false, nil
	}
	defer win.CoTaskMemFree(pidl)

	dlg.FilePath, err = pathFromPIDL(pidl)
	accepted = dlg.FilePath != ""
	return
}
