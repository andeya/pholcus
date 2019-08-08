// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"bytes"
)

import (
	"github.com/lxn/win"
)

type Key uint16

func (k Key) String() string {
	return key2string[k]
}

const (
	KeyLButton           Key = win.VK_LBUTTON
	KeyRButton           Key = win.VK_RBUTTON
	KeyCancel            Key = win.VK_CANCEL
	KeyMButton           Key = win.VK_MBUTTON
	KeyXButton1          Key = win.VK_XBUTTON1
	KeyXButton2          Key = win.VK_XBUTTON2
	KeyBack              Key = win.VK_BACK
	KeyTab               Key = win.VK_TAB
	KeyClear             Key = win.VK_CLEAR
	KeyReturn            Key = win.VK_RETURN
	KeyShift             Key = win.VK_SHIFT
	KeyControl           Key = win.VK_CONTROL
	KeyAlt               Key = win.VK_MENU
	KeyMenu              Key = win.VK_MENU
	KeyPause             Key = win.VK_PAUSE
	KeyCapital           Key = win.VK_CAPITAL
	KeyKana              Key = win.VK_KANA
	KeyHangul            Key = win.VK_HANGUL
	KeyJunja             Key = win.VK_JUNJA
	KeyFinal             Key = win.VK_FINAL
	KeyHanja             Key = win.VK_HANJA
	KeyKanji             Key = win.VK_KANJI
	KeyEscape            Key = win.VK_ESCAPE
	KeyConvert           Key = win.VK_CONVERT
	KeyNonconvert        Key = win.VK_NONCONVERT
	KeyAccept            Key = win.VK_ACCEPT
	KeyModeChange        Key = win.VK_MODECHANGE
	KeySpace             Key = win.VK_SPACE
	KeyPrior             Key = win.VK_PRIOR
	KeyNext              Key = win.VK_NEXT
	KeyEnd               Key = win.VK_END
	KeyHome              Key = win.VK_HOME
	KeyLeft              Key = win.VK_LEFT
	KeyUp                Key = win.VK_UP
	KeyRight             Key = win.VK_RIGHT
	KeyDown              Key = win.VK_DOWN
	KeySelect            Key = win.VK_SELECT
	KeyPrint             Key = win.VK_PRINT
	KeyExecute           Key = win.VK_EXECUTE
	KeySnapshot          Key = win.VK_SNAPSHOT
	KeyInsert            Key = win.VK_INSERT
	KeyDelete            Key = win.VK_DELETE
	KeyHelp              Key = win.VK_HELP
	Key0                 Key = 0x30
	Key1                 Key = 0x31
	Key2                 Key = 0x32
	Key3                 Key = 0x33
	Key4                 Key = 0x34
	Key5                 Key = 0x35
	Key6                 Key = 0x36
	Key7                 Key = 0x37
	Key8                 Key = 0x38
	Key9                 Key = 0x39
	KeyA                 Key = 0x41
	KeyB                 Key = 0x42
	KeyC                 Key = 0x43
	KeyD                 Key = 0x44
	KeyE                 Key = 0x45
	KeyF                 Key = 0x46
	KeyG                 Key = 0x47
	KeyH                 Key = 0x48
	KeyI                 Key = 0x49
	KeyJ                 Key = 0x4A
	KeyK                 Key = 0x4B
	KeyL                 Key = 0x4C
	KeyM                 Key = 0x4D
	KeyN                 Key = 0x4E
	KeyO                 Key = 0x4F
	KeyP                 Key = 0x50
	KeyQ                 Key = 0x51
	KeyR                 Key = 0x52
	KeyS                 Key = 0x53
	KeyT                 Key = 0x54
	KeyU                 Key = 0x55
	KeyV                 Key = 0x56
	KeyW                 Key = 0x57
	KeyX                 Key = 0x58
	KeyY                 Key = 0x59
	KeyZ                 Key = 0x5A
	KeyLWin              Key = win.VK_LWIN
	KeyRWin              Key = win.VK_RWIN
	KeyApps              Key = win.VK_APPS
	KeySleep             Key = win.VK_SLEEP
	KeyNumpad0           Key = win.VK_NUMPAD0
	KeyNumpad1           Key = win.VK_NUMPAD1
	KeyNumpad2           Key = win.VK_NUMPAD2
	KeyNumpad3           Key = win.VK_NUMPAD3
	KeyNumpad4           Key = win.VK_NUMPAD4
	KeyNumpad5           Key = win.VK_NUMPAD5
	KeyNumpad6           Key = win.VK_NUMPAD6
	KeyNumpad7           Key = win.VK_NUMPAD7
	KeyNumpad8           Key = win.VK_NUMPAD8
	KeyNumpad9           Key = win.VK_NUMPAD9
	KeyMultiply          Key = win.VK_MULTIPLY
	KeyAdd               Key = win.VK_ADD
	KeySeparator         Key = win.VK_SEPARATOR
	KeySubtract          Key = win.VK_SUBTRACT
	KeyDecimal           Key = win.VK_DECIMAL
	KeyDivide            Key = win.VK_DIVIDE
	KeyF1                Key = win.VK_F1
	KeyF2                Key = win.VK_F2
	KeyF3                Key = win.VK_F3
	KeyF4                Key = win.VK_F4
	KeyF5                Key = win.VK_F5
	KeyF6                Key = win.VK_F6
	KeyF7                Key = win.VK_F7
	KeyF8                Key = win.VK_F8
	KeyF9                Key = win.VK_F9
	KeyF10               Key = win.VK_F10
	KeyF11               Key = win.VK_F11
	KeyF12               Key = win.VK_F12
	KeyF13               Key = win.VK_F13
	KeyF14               Key = win.VK_F14
	KeyF15               Key = win.VK_F15
	KeyF16               Key = win.VK_F16
	KeyF17               Key = win.VK_F17
	KeyF18               Key = win.VK_F18
	KeyF19               Key = win.VK_F19
	KeyF20               Key = win.VK_F20
	KeyF21               Key = win.VK_F21
	KeyF22               Key = win.VK_F22
	KeyF23               Key = win.VK_F23
	KeyF24               Key = win.VK_F24
	KeyNumlock           Key = win.VK_NUMLOCK
	KeyScroll            Key = win.VK_SCROLL
	KeyLShift            Key = win.VK_LSHIFT
	KeyRShift            Key = win.VK_RSHIFT
	KeyLControl          Key = win.VK_LCONTROL
	KeyRControl          Key = win.VK_RCONTROL
	KeyLAlt              Key = win.VK_LMENU
	KeyLMenu             Key = win.VK_LMENU
	KeyRAlt              Key = win.VK_RMENU
	KeyRMenu             Key = win.VK_RMENU
	KeyBrowserBack       Key = win.VK_BROWSER_BACK
	KeyBrowserForward    Key = win.VK_BROWSER_FORWARD
	KeyBrowserRefresh    Key = win.VK_BROWSER_REFRESH
	KeyBrowserStop       Key = win.VK_BROWSER_STOP
	KeyBrowserSearch     Key = win.VK_BROWSER_SEARCH
	KeyBrowserFavorites  Key = win.VK_BROWSER_FAVORITES
	KeyBrowserHome       Key = win.VK_BROWSER_HOME
	KeyVolumeMute        Key = win.VK_VOLUME_MUTE
	KeyVolumeDown        Key = win.VK_VOLUME_DOWN
	KeyVolumeUp          Key = win.VK_VOLUME_UP
	KeyMediaNextTrack    Key = win.VK_MEDIA_NEXT_TRACK
	KeyMediaPrevTrack    Key = win.VK_MEDIA_PREV_TRACK
	KeyMediaStop         Key = win.VK_MEDIA_STOP
	KeyMediaPlayPause    Key = win.VK_MEDIA_PLAY_PAUSE
	KeyLaunchMail        Key = win.VK_LAUNCH_MAIL
	KeyLaunchMediaSelect Key = win.VK_LAUNCH_MEDIA_SELECT
	KeyLaunchApp1        Key = win.VK_LAUNCH_APP1
	KeyLaunchApp2        Key = win.VK_LAUNCH_APP2
	KeyOEM1              Key = win.VK_OEM_1
	KeyOEMPlus           Key = win.VK_OEM_PLUS
	KeyOEMComma          Key = win.VK_OEM_COMMA
	KeyOEMMinus          Key = win.VK_OEM_MINUS
	KeyOEMPeriod         Key = win.VK_OEM_PERIOD
	KeyOEM2              Key = win.VK_OEM_2
	KeyOEM3              Key = win.VK_OEM_3
	KeyOEM4              Key = win.VK_OEM_4
	KeyOEM5              Key = win.VK_OEM_5
	KeyOEM6              Key = win.VK_OEM_6
	KeyOEM7              Key = win.VK_OEM_7
	KeyOEM8              Key = win.VK_OEM_8
	KeyOEM102            Key = win.VK_OEM_102
	KeyProcessKey        Key = win.VK_PROCESSKEY
	KeyPacket            Key = win.VK_PACKET
	KeyAttn              Key = win.VK_ATTN
	KeyCRSel             Key = win.VK_CRSEL
	KeyEXSel             Key = win.VK_EXSEL
	KeyErEOF             Key = win.VK_EREOF
	KeyPlay              Key = win.VK_PLAY
	KeyZoom              Key = win.VK_ZOOM
	KeyNoName            Key = win.VK_NONAME
	KeyPA1               Key = win.VK_PA1
	KeyOEMClear          Key = win.VK_OEM_CLEAR
)

var key2string = map[Key]string{
	KeyLButton:           "LButton",
	KeyRButton:           "RButton",
	KeyCancel:            "Cancel",
	KeyMButton:           "MButton",
	KeyXButton1:          "XButton1",
	KeyXButton2:          "XButton2",
	KeyBack:              "Back",
	KeyTab:               "Tab",
	KeyClear:             "Clear",
	KeyReturn:            "Return",
	KeyShift:             "Shift",
	KeyControl:           "Control",
	KeyAlt:               "Alt / Menu",
	KeyPause:             "Pause",
	KeyCapital:           "Capital",
	KeyKana:              "Kana / Hangul",
	KeyJunja:             "Junja",
	KeyFinal:             "Final",
	KeyHanja:             "Hanja / Kanji",
	KeyEscape:            "Escape",
	KeyConvert:           "Convert",
	KeyNonconvert:        "Nonconvert",
	KeyAccept:            "Accept",
	KeyModeChange:        "ModeChange",
	KeySpace:             "Space",
	KeyPrior:             "Prior",
	KeyNext:              "Next",
	KeyEnd:               "End",
	KeyHome:              "Home",
	KeyLeft:              "Left",
	KeyUp:                "Up",
	KeyRight:             "Right",
	KeyDown:              "Down",
	KeySelect:            "Select",
	KeyPrint:             "Print",
	KeyExecute:           "Execute",
	KeySnapshot:          "Snapshot",
	KeyInsert:            "Insert",
	KeyDelete:            "Delete",
	KeyHelp:              "Help",
	Key0:                 "0",
	Key1:                 "1",
	Key2:                 "2",
	Key3:                 "3",
	Key4:                 "4",
	Key5:                 "5",
	Key6:                 "6",
	Key7:                 "7",
	Key8:                 "8",
	Key9:                 "9",
	KeyA:                 "A",
	KeyB:                 "B",
	KeyC:                 "C",
	KeyD:                 "D",
	KeyE:                 "E",
	KeyF:                 "F",
	KeyG:                 "G",
	KeyH:                 "H",
	KeyI:                 "I",
	KeyJ:                 "J",
	KeyK:                 "K",
	KeyL:                 "L",
	KeyM:                 "M",
	KeyN:                 "N",
	KeyO:                 "O",
	KeyP:                 "P",
	KeyQ:                 "Q",
	KeyR:                 "R",
	KeyS:                 "S",
	KeyT:                 "T",
	KeyU:                 "U",
	KeyV:                 "V",
	KeyW:                 "W",
	KeyX:                 "X",
	KeyY:                 "Y",
	KeyZ:                 "Z",
	KeyLWin:              "LWin",
	KeyRWin:              "RWin",
	KeyApps:              "Apps",
	KeySleep:             "Sleep",
	KeyNumpad0:           "Numpad0",
	KeyNumpad1:           "Numpad1",
	KeyNumpad2:           "Numpad2",
	KeyNumpad3:           "Numpad3",
	KeyNumpad4:           "Numpad4",
	KeyNumpad5:           "Numpad5",
	KeyNumpad6:           "Numpad6",
	KeyNumpad7:           "Numpad7",
	KeyNumpad8:           "Numpad8",
	KeyNumpad9:           "Numpad9",
	KeyMultiply:          "Multiply",
	KeyAdd:               "Add",
	KeySeparator:         "Separator",
	KeySubtract:          "Subtract",
	KeyDecimal:           "Decimal",
	KeyDivide:            "Divide",
	KeyF1:                "F1",
	KeyF2:                "F2",
	KeyF3:                "F3",
	KeyF4:                "F4",
	KeyF5:                "F5",
	KeyF6:                "F6",
	KeyF7:                "F7",
	KeyF8:                "F8",
	KeyF9:                "F9",
	KeyF10:               "F10",
	KeyF11:               "F11",
	KeyF12:               "F12",
	KeyF13:               "F13",
	KeyF14:               "F14",
	KeyF15:               "F15",
	KeyF16:               "F16",
	KeyF17:               "F17",
	KeyF18:               "F18",
	KeyF19:               "F19",
	KeyF20:               "F20",
	KeyF21:               "F21",
	KeyF22:               "F22",
	KeyF23:               "F23",
	KeyF24:               "F24",
	KeyNumlock:           "Numlock",
	KeyScroll:            "Scroll",
	KeyLShift:            "LShift",
	KeyRShift:            "RShift",
	KeyLControl:          "LControl",
	KeyRControl:          "RControl",
	KeyLMenu:             "LMenu",
	KeyRMenu:             "RMenu",
	KeyBrowserBack:       "BrowserBack",
	KeyBrowserForward:    "BrowserForward",
	KeyBrowserRefresh:    "BrowserRefresh",
	KeyBrowserStop:       "BrowserStop",
	KeyBrowserSearch:     "BrowserSearch",
	KeyBrowserFavorites:  "BrowserFavorites",
	KeyBrowserHome:       "BrowserHome",
	KeyVolumeMute:        "VolumeMute",
	KeyVolumeDown:        "VolumeDown",
	KeyVolumeUp:          "VolumeUp",
	KeyMediaNextTrack:    "MediaNextTrack",
	KeyMediaPrevTrack:    "MediaPrevTrack",
	KeyMediaStop:         "MediaStop",
	KeyMediaPlayPause:    "MediaPlayPause",
	KeyLaunchMail:        "LaunchMail",
	KeyLaunchMediaSelect: "LaunchMediaSelect",
	KeyLaunchApp1:        "LaunchApp1",
	KeyLaunchApp2:        "LaunchApp2",
	KeyOEM1:              "OEM1",
	KeyOEMPlus:           "OEMPlus",
	KeyOEMComma:          "OEMComma",
	KeyOEMMinus:          "OEMMinus",
	KeyOEMPeriod:         "OEMPeriod",
	KeyOEM2:              "OEM2",
	KeyOEM3:              "OEM3",
	KeyOEM4:              "OEM4",
	KeyOEM5:              "OEM5",
	KeyOEM6:              "OEM6",
	KeyOEM7:              "OEM7",
	KeyOEM8:              "OEM8",
	KeyOEM102:            "OEM102",
	KeyProcessKey:        "ProcessKey",
	KeyPacket:            "Packet",
	KeyAttn:              "Attn",
	KeyCRSel:             "CRSel",
	KeyEXSel:             "EXSel",
	KeyErEOF:             "ErEOF",
	KeyPlay:              "Play",
	KeyZoom:              "Zoom",
	KeyNoName:            "NoName",
	KeyPA1:               "PA1",
	KeyOEMClear:          "OEMClear",
}

type Modifiers byte

func (m Modifiers) String() string {
	return modifiers2string[m]
}

var modifiers2string = map[Modifiers]string{
	ModShift:                       "Shift",
	ModControl:                     "Ctrl",
	ModControl | ModShift:          "Ctrl+Shift",
	ModAlt:                         "Alt",
	ModAlt | ModShift:              "Alt+Shift",
	ModAlt | ModControl:            "Alt+Ctrl",
	ModAlt | ModControl | ModShift: "Alt+Ctrl+Shift",
}

const (
	ModShift Modifiers = 1 << iota
	ModControl
	ModAlt
)

func ModifiersDown() Modifiers {
	var m Modifiers

	if ShiftDown() {
		m |= ModShift
	}
	if ControlDown() {
		m |= ModControl
	}
	if AltDown() {
		m |= ModAlt
	}

	return m
}

type Shortcut struct {
	Modifiers Modifiers
	Key       Key
}

func (s Shortcut) String() string {
	m := s.Modifiers.String()
	if m == "" {
		return s.Key.String()
	}

	b := new(bytes.Buffer)

	b.WriteString(m)
	b.WriteRune('+')
	b.WriteString(s.Key.String())

	return b.String()
}

func AltDown() bool {
	return win.GetKeyState(int32(KeyAlt))>>15 != 0
}

func ControlDown() bool {
	return win.GetKeyState(int32(KeyControl))>>15 != 0
}

func ShiftDown() bool {
	return win.GetKeyState(int32(KeyShift))>>15 != 0
}
