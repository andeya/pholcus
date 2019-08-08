// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type Alignment1D uint

const (
	AlignDefault Alignment1D = iota
	AlignNear
	AlignCenter
	AlignFar
)

type Alignment2D uint

const (
	AlignHVDefault Alignment2D = iota
	AlignHNearVNear
	AlignHCenterVNear
	AlignHFarVNear
	AlignHNearVCenter
	AlignHCenterVCenter
	AlignHFarVCenter
	AlignHNearVFar
	AlignHCenterVFar
	AlignHFarVFar
)
