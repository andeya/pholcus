// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type Alignment1D uint

const (
	AlignNear Alignment1D = iota
	AlignCenter
	AlignFar
)

type Alignment2D uint

const (
	AlignHNearVNear Alignment2D = iota
	AlignHCenterVNear
	AlignHFarVNear
	AlignHNearVCenter
	AlignHCenterVCenter
	AlignHFarVCenter
	AlignHNearVFar
	AlignHCenterVFar
	AlignHFarVFar
)
