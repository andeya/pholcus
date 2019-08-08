// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"strings"
	"syscall"
)

import (
	"github.com/lxn/win"
)

type MsgBoxStyle uint

const (
	MsgBoxOK                  MsgBoxStyle = win.MB_OK
	MsgBoxOKCancel            MsgBoxStyle = win.MB_OKCANCEL
	MsgBoxAbortRetryIgnore    MsgBoxStyle = win.MB_ABORTRETRYIGNORE
	MsgBoxYesNoCancel         MsgBoxStyle = win.MB_YESNOCANCEL
	MsgBoxYesNo               MsgBoxStyle = win.MB_YESNO
	MsgBoxRetryCancel         MsgBoxStyle = win.MB_RETRYCANCEL
	MsgBoxCancelTryContinue   MsgBoxStyle = win.MB_CANCELTRYCONTINUE
	MsgBoxIconHand            MsgBoxStyle = win.MB_ICONHAND
	MsgBoxIconQuestion        MsgBoxStyle = win.MB_ICONQUESTION
	MsgBoxIconExclamation     MsgBoxStyle = win.MB_ICONEXCLAMATION
	MsgBoxIconAsterisk        MsgBoxStyle = win.MB_ICONASTERISK
	MsgBoxUserIcon            MsgBoxStyle = win.MB_USERICON
	MsgBoxIconWarning         MsgBoxStyle = win.MB_ICONWARNING
	MsgBoxIconError           MsgBoxStyle = win.MB_ICONERROR
	MsgBoxIconInformation     MsgBoxStyle = win.MB_ICONINFORMATION
	MsgBoxIconStop            MsgBoxStyle = win.MB_ICONSTOP
	MsgBoxDefButton1          MsgBoxStyle = win.MB_DEFBUTTON1
	MsgBoxDefButton2          MsgBoxStyle = win.MB_DEFBUTTON2
	MsgBoxDefButton3          MsgBoxStyle = win.MB_DEFBUTTON3
	MsgBoxDefButton4          MsgBoxStyle = win.MB_DEFBUTTON4
	MsgBoxApplModal           MsgBoxStyle = win.MB_APPLMODAL
	MsgBoxSystemModal         MsgBoxStyle = win.MB_SYSTEMMODAL
	MsgBoxTaskModal           MsgBoxStyle = win.MB_TASKMODAL
	MsgBoxHelp                MsgBoxStyle = win.MB_HELP
	MsgBoxSetForeground       MsgBoxStyle = win.MB_SETFOREGROUND
	MsgBoxDefaultDesktopOnly  MsgBoxStyle = win.MB_DEFAULT_DESKTOP_ONLY
	MsgBoxTopMost             MsgBoxStyle = win.MB_TOPMOST
	MsgBoxRight               MsgBoxStyle = win.MB_RIGHT
	MsgBoxRTLReading          MsgBoxStyle = win.MB_RTLREADING
	MsgBoxServiceNotification MsgBoxStyle = win.MB_SERVICE_NOTIFICATION
)

func MsgBox(owner Form, title, message string, style MsgBoxStyle) int {
	var ownerHWnd win.HWND

	if owner != nil {
		ownerHWnd = owner.Handle()
	}

	return int(win.MessageBox(
		ownerHWnd,
		syscall.StringToUTF16Ptr(strings.ReplaceAll(message, "\x00", "␀")),
		syscall.StringToUTF16Ptr(strings.ReplaceAll(title, "\x00", "␀")),
		uint32(style)))
}
