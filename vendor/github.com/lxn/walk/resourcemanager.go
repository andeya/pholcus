// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func init() {
	Resources.rootDirPath, _ = os.Getwd()
	Resources.bitmaps = make(map[string]*Bitmap)
	Resources.icons = make(map[string]*Icon)
}

// Resources is the singleton instance of ResourceManager.
var Resources ResourceManager

// ResourceManager is a cache for sharing resources like bitmaps and icons.
// The resources can be either embedded in the running executable
// file or located below a specified root directory in the file system.
type ResourceManager struct {
	rootDirPath string
	bitmaps     map[string]*Bitmap
	icons       map[string]*Icon
}

// RootDirPath returns the root directory path where resources are to be loaded from.
func (rm *ResourceManager) RootDirPath() string {
	return rm.rootDirPath
}

// SetRootDirPath sets the root directory path where resources are to be loaded from.
func (rm *ResourceManager) SetRootDirPath(rootDirPath string) error {
	path, err := filepath.Abs(rootDirPath)
	if err != nil {
		return err
	}

	rm.rootDirPath = path

	return nil
}

// Bitmap returns the Bitmap identified by name, or an error if it could not be found.
func (rm *ResourceManager) Bitmap(name string) (*Bitmap, error) {
	if bm := rm.bitmaps[name]; bm != nil {
		return bm, nil
	}

	if bm, err := NewBitmapFromFile(filepath.Join(rm.rootDirPath, name)); err == nil {
		rm.bitmaps[name] = bm
		return bm, nil
	}

	if bm, err := NewBitmapFromResource(name); err == nil {
		rm.bitmaps[name] = bm
		return bm, nil
	}

	if id, err := strconv.Atoi(name); err == nil {
		if bm, err := NewBitmapFromResourceId(id); err == nil {
			rm.bitmaps[name] = bm
			return bm, nil
		}
	}

	return nil, rm.notFoundErr("bitmap", name)
}

// Icon returns the Icon identified by name, or an error if it could not be found.
func (rm *ResourceManager) Icon(name string) (*Icon, error) {
	if icon := rm.icons[name]; icon != nil {
		return icon, nil
	}

	if icon, err := NewIconFromFile(filepath.Join(rm.rootDirPath, name)); err == nil {
		rm.icons[name] = icon
		return icon, nil
	}

	if icon, err := NewIconFromResource(name); err == nil {
		rm.icons[name] = icon
		return icon, nil
	}

	if id, err := strconv.Atoi(name); err == nil {
		if icon, err := NewIconFromResourceId(id); err == nil {
			rm.icons[name] = icon
			return icon, nil
		}
	}

	return nil, rm.notFoundErr("icon", name)
}

// Image returns the Image identified by name, or an error if it could not be found.
func (rm *ResourceManager) Image(name string) (Image, error) {
	if icon, err := rm.Icon(name); err == nil {
		return icon, nil
	}

	if bm, err := rm.Bitmap(name); err == nil {
		return bm, nil
	}

	return nil, rm.notFoundErr("image", name)
}

func (rm *ResourceManager) notFoundErr(typ, name string) error {
	path := filepath.Clean(filepath.Join(rm.rootDirPath, name))

	return newError(fmt.Sprintf("neither %s resource '%s' nor file '%s' could be found or the image format is not supported", typ, name, path))
}
