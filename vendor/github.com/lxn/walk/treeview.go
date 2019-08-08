// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

type treeViewItemInfo struct {
	handle       win.HTREEITEM
	child2Handle map[TreeItem]win.HTREEITEM
}

type TreeView struct {
	WidgetBase
	model                         TreeModel
	lazyPopulation                bool
	itemsResetEventHandlerHandle  int
	itemChangedEventHandlerHandle int
	item2Info                     map[TreeItem]*treeViewItemInfo
	handle2Item                   map[win.HTREEITEM]TreeItem
	currItem                      TreeItem
	hIml                          win.HIMAGELIST
	usingSysIml                   bool
	imageUintptr2Index            map[uintptr]int32
	filePath2IconIndex            map[string]int32
	expandedChangedPublisher      TreeItemEventPublisher
	currentItemChangedPublisher   EventPublisher
	itemActivatedPublisher        EventPublisher
}

func NewTreeView(parent Container) (*TreeView, error) {
	tv := new(TreeView)

	if err := InitWidget(
		tv,
		parent,
		"SysTreeView32",
		win.WS_TABSTOP|win.WS_VISIBLE|win.TVS_HASBUTTONS|win.TVS_LINESATROOT|win.TVS_SHOWSELALWAYS|win.TVS_TRACKSELECT,
		win.WS_EX_CLIENTEDGE); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			tv.Dispose()
		}
	}()

	if hr := win.HRESULT(tv.SendMessage(win.TVM_SETEXTENDEDSTYLE, win.TVS_EX_DOUBLEBUFFER, win.TVS_EX_DOUBLEBUFFER)); win.FAILED(hr) {
		return nil, errorFromHRESULT("TVM_SETEXTENDEDSTYLE", hr)
	}

	if err := tv.setTheme("Explorer"); err != nil {
		return nil, err
	}

	tv.GraphicsEffects().Add(InteractionEffect)
	tv.GraphicsEffects().Add(FocusEffect)

	tv.MustRegisterProperty("CurrentItem", NewReadOnlyProperty(
		func() interface{} {
			return tv.CurrentItem()
		},
		tv.CurrentItemChanged()))

	tv.MustRegisterProperty("CurrentItemLevel", NewReadOnlyProperty(
		func() interface{} {
			level := -1
			item := tv.CurrentItem()

			for item != nil {
				level++
				item = item.Parent()
			}

			return level
		},
		tv.CurrentItemChanged()))

	tv.MustRegisterProperty("HasCurrentItem", NewReadOnlyBoolProperty(
		func() bool {
			return tv.CurrentItem() != nil
		},
		tv.CurrentItemChanged()))

	succeeded = true

	return tv, nil
}

func (*TreeView) LayoutFlags() LayoutFlags {
	return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert | GreedyHorz | GreedyVert
}

func (tv *TreeView) SizeHint() Size {
	return tv.dialogBaseUnitsToPixels(Size{100, 100})
}

func (tv *TreeView) Dispose() {
	tv.WidgetBase.Dispose()

	tv.disposeImageListAndCaches()
}

func (tv *TreeView) SetBackground(bg Brush) {
	tv.WidgetBase.SetBackground(bg)

	color := Color(win.GetSysColor(win.COLOR_WINDOW))

	if bg != nil {
		type Colorer interface {
			Color() Color
		}

		if c, ok := bg.(Colorer); ok {
			color = c.Color()
		}
	}

	tv.SendMessage(win.TVM_SETBKCOLOR, 0, uintptr(color))
}

func (tv *TreeView) Model() TreeModel {
	return tv.model
}

func (tv *TreeView) SetModel(model TreeModel) error {
	if tv.model != nil {
		tv.model.ItemsReset().Detach(tv.itemsResetEventHandlerHandle)
		tv.model.ItemChanged().Detach(tv.itemChangedEventHandlerHandle)

		tv.disposeImageListAndCaches()
	}

	tv.model = model

	if model != nil {
		tv.lazyPopulation = model.LazyPopulation()

		tv.itemsResetEventHandlerHandle = model.ItemsReset().Attach(func(parent TreeItem) {
			if parent == nil {
				tv.resetItems()
			} else if tv.item2Info[parent] != nil {
				tv.SetSuspended(true)
				defer tv.SetSuspended(false)

				if err := tv.removeDescendants(parent); err != nil {
					return
				}

				if err := tv.insertChildren(parent); err != nil {
					return
				}
			}
		})

		tv.itemChangedEventHandlerHandle = model.ItemChanged().Attach(func(item TreeItem) {
			if item == nil || tv.item2Info[item] == nil {
				return
			}

			if err := tv.updateItem(item); err != nil {
				return
			}
		})
	}

	return tv.resetItems()
}

func (tv *TreeView) CurrentItem() TreeItem {
	return tv.currItem
}

func (tv *TreeView) SetCurrentItem(item TreeItem) error {
	if item == tv.currItem {
		return nil
	}

	if item != nil {
		if err := tv.ensureItemAndAncestorsInserted(item); err != nil {
			return err
		}
	}

	var handle win.HTREEITEM
	if item != nil {
		if info := tv.item2Info[item]; info == nil {
			return newError("invalid item")
		} else {
			handle = info.handle
		}
	}

	if 0 == tv.SendMessage(win.TVM_SELECTITEM, win.TVGN_CARET, uintptr(handle)) {
		return newError("SendMessage(TVM_SELECTITEM) failed")
	}

	tv.currItem = item

	return nil
}

func (tv *TreeView) ItemAt(x, y int) TreeItem {
	hti := win.TVHITTESTINFO{Pt: win.POINT{int32(x), int32(y)}}

	tv.SendMessage(win.TVM_HITTEST, 0, uintptr(unsafe.Pointer(&hti)))

	if item, ok := tv.handle2Item[hti.HItem]; ok {
		return item
	}

	return nil
}

func (tv *TreeView) ItemHeight() int {
	return int(tv.SendMessage(win.TVM_GETITEMHEIGHT, 0, 0))
}

func (tv *TreeView) SetItemHeight(height int) {
	tv.SendMessage(win.TVM_SETITEMHEIGHT, uintptr(height), 0)
}

func (tv *TreeView) resetItems() error {
	tv.SetSuspended(true)
	defer tv.SetSuspended(false)

	if err := tv.clearItems(); err != nil {
		return err
	}

	if tv.model == nil {
		return nil
	}

	if err := tv.insertRoots(); err != nil {
		return err
	}

	return nil
}

func (tv *TreeView) clearItems() error {
	if 0 == tv.SendMessage(win.TVM_DELETEITEM, 0, 0) {
		return newError("SendMessage(TVM_DELETEITEM) failed")
	}

	tv.item2Info = make(map[TreeItem]*treeViewItemInfo)
	tv.handle2Item = make(map[win.HTREEITEM]TreeItem)

	return nil
}

func (tv *TreeView) insertRoots() error {
	for i := tv.model.RootCount() - 1; i >= 0; i-- {
		if _, err := tv.insertItem(i, tv.model.RootAt(i)); err != nil {
			return err
		}
	}

	return nil
}

func (tv *TreeView) ApplyDPI(dpi int) {
	tv.WidgetBase.ApplyDPI(dpi)

	tv.disposeImageListAndCaches()
}

func (tv *TreeView) applyImageListForImage(image interface{}) {
	tv.hIml, tv.usingSysIml, _ = imageListForImage(image, tv.DPI())

	tv.SendMessage(win.TVM_SETIMAGELIST, 0, uintptr(tv.hIml))

	tv.imageUintptr2Index = make(map[uintptr]int32)
	tv.filePath2IconIndex = make(map[string]int32)
}

func (tv *TreeView) disposeImageListAndCaches() {
	if tv.hIml != 0 && !tv.usingSysIml {
		win.ImageList_Destroy(tv.hIml)
	}
	tv.hIml = 0

	tv.imageUintptr2Index = nil
	tv.filePath2IconIndex = nil
}

func (tv *TreeView) setTVITEMImageInfo(tvi *win.TVITEM, item TreeItem) {
	if imager, ok := item.(Imager); ok {
		if tv.hIml == 0 {
			tv.applyImageListForImage(imager.Image())
		}

		// FIXME: If not setting TVIF_SELECTEDIMAGE and tvi.ISelectedImage,
		// some default icon will show up, even though we have not asked for it.

		tvi.Mask |= win.TVIF_IMAGE | win.TVIF_SELECTEDIMAGE
		tvi.IImage = imageIndexMaybeAdd(
			imager.Image(),
			tv.hIml,
			tv.usingSysIml,
			tv.imageUintptr2Index,
			tv.filePath2IconIndex,
			tv.DPI())

		tvi.ISelectedImage = tvi.IImage
	}
}

func (tv *TreeView) insertItem(index int, item TreeItem) (win.HTREEITEM, error) {
	var tvins win.TVINSERTSTRUCT
	tvi := &tvins.Item

	tvi.Mask = win.TVIF_CHILDREN | win.TVIF_TEXT
	tvi.PszText = win.LPSTR_TEXTCALLBACK
	tvi.CChildren = win.I_CHILDRENCALLBACK

	tv.setTVITEMImageInfo(tvi, item)

	parent := item.Parent()

	if parent == nil {
		tvins.HParent = win.TVI_ROOT
	} else {
		info := tv.item2Info[parent]
		if info == nil {
			return 0, newError("invalid parent")
		}
		tvins.HParent = info.handle
	}

	tvins.HInsertAfter = win.TVI_FIRST

	hItem := win.HTREEITEM(tv.SendMessage(win.TVM_INSERTITEM, 0, uintptr(unsafe.Pointer(&tvins))))
	if hItem == 0 {
		return 0, newError("TVM_INSERTITEM failed")
	}
	tv.item2Info[item] = &treeViewItemInfo{hItem, make(map[TreeItem]win.HTREEITEM)}
	tv.handle2Item[hItem] = item

	if !tv.lazyPopulation {
		if err := tv.insertChildren(item); err != nil {
			return 0, err
		}
	}

	return hItem, nil
}

func (tv *TreeView) insertChildren(parent TreeItem) error {
	info := tv.item2Info[parent]

	for i := parent.ChildCount() - 1; i >= 0; i-- {
		child := parent.ChildAt(i)

		if handle, err := tv.insertItem(i, child); err != nil {
			return err
		} else {
			info.child2Handle[child] = handle
		}
	}

	return nil
}

func (tv *TreeView) updateItem(item TreeItem) error {
	tvi := &win.TVITEM{
		Mask:    win.TVIF_TEXT,
		HItem:   tv.item2Info[item].handle,
		PszText: win.LPSTR_TEXTCALLBACK,
	}

	tv.setTVITEMImageInfo(tvi, item)

	if 0 == tv.SendMessage(win.TVM_SETITEM, 0, uintptr(unsafe.Pointer(tvi))) {
		return newError("SendMessage(TVM_SETITEM) failed")
	}

	return nil
}

func (tv *TreeView) removeItem(item TreeItem) error {
	if err := tv.removeDescendants(item); err != nil {
		return err
	}

	info := tv.item2Info[item]
	if info == nil {
		return newError("invalid item")
	}

	if 0 == tv.SendMessage(win.TVM_DELETEITEM, 0, uintptr(info.handle)) {
		return newError("SendMessage(TVM_DELETEITEM) failed")
	}

	if parentInfo := tv.item2Info[item.Parent()]; parentInfo != nil {
		delete(parentInfo.child2Handle, item)
	}
	delete(tv.item2Info, item)
	delete(tv.handle2Item, info.handle)

	return nil
}

func (tv *TreeView) removeDescendants(parent TreeItem) error {
	for item, _ := range tv.item2Info[parent].child2Handle {
		if err := tv.removeItem(item); err != nil {
			return err
		}
	}

	return nil
}

func (tv *TreeView) ensureItemAndAncestorsInserted(item TreeItem) error {
	if item == nil {
		return newError("invalid item")
	}

	tv.SetSuspended(true)
	defer tv.SetSuspended(false)

	var hierarchy []TreeItem

	for item != nil && tv.item2Info[item] == nil {
		item = item.Parent()

		if item != nil {
			hierarchy = append(hierarchy, item)
		} else {
			return newError("invalid item")
		}
	}

	for i := len(hierarchy) - 1; i >= 0; i-- {
		if err := tv.insertChildren(hierarchy[i]); err != nil {
			return err
		}
	}

	return nil
}

func (tv *TreeView) Expanded(item TreeItem) bool {
	if tv.item2Info[item] == nil {
		return false
	}

	tvi := &win.TVITEM{
		HItem:     tv.item2Info[item].handle,
		Mask:      win.TVIF_STATE,
		StateMask: win.TVIS_EXPANDED,
	}

	if 0 == tv.SendMessage(win.TVM_GETITEM, 0, uintptr(unsafe.Pointer(tvi))) {
		newError("SendMessage(TVM_GETITEM) failed")
	}

	return tvi.State&win.TVIS_EXPANDED != 0
}

func (tv *TreeView) SetExpanded(item TreeItem, expanded bool) error {
	if expanded {
		if err := tv.ensureItemAndAncestorsInserted(item); err != nil {
			return err
		}
	}

	info := tv.item2Info[item]
	if info == nil {
		return newError("invalid item")
	}

	var action uintptr
	if expanded {
		action = win.TVE_EXPAND
	} else {
		action = win.TVE_COLLAPSE
	}

	if 0 == tv.SendMessage(win.TVM_EXPAND, action, uintptr(info.handle)) {
		return newError("SendMessage(TVM_EXPAND) failed")
	}

	return nil
}

func (tv *TreeView) ExpandedChanged() *TreeItemEvent {
	return tv.expandedChangedPublisher.Event()
}

func (tv *TreeView) CurrentItemChanged() *Event {
	return tv.currentItemChangedPublisher.Event()
}

func (tv *TreeView) ItemActivated() *Event {
	return tv.itemActivatedPublisher.Event()
}

func (tv *TreeView) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_GETDLGCODE:
		if wParam == win.VK_RETURN {
			return win.DLGC_WANTALLKEYS
		}

	case win.WM_NOTIFY:
		nmhdr := (*win.NMHDR)(unsafe.Pointer(lParam))

		switch nmhdr.Code {
		case win.TVN_GETDISPINFO:
			nmtvdi := (*win.NMTVDISPINFO)(unsafe.Pointer(lParam))
			item := tv.handle2Item[nmtvdi.Item.HItem]

			if nmtvdi.Item.Mask&win.TVIF_TEXT != 0 {
				var text string
				rc := win.RECT{Left: int32(nmtvdi.Item.HItem)}
				if 0 != tv.SendMessage(win.TVM_GETITEMRECT, 0, uintptr(unsafe.Pointer(&rc))) {
					// Only retrieve text if the item is visible. Why isn't Windows doing this for us?
					text = item.Text()
				}

				utf16 := syscall.StringToUTF16(text)
				buf := (*[264]uint16)(unsafe.Pointer(nmtvdi.Item.PszText))
				max := mini(len(utf16), int(nmtvdi.Item.CchTextMax))
				copy((*buf)[:], utf16[:max])
				(*buf)[max-1] = 0
			}
			if nmtvdi.Item.Mask&win.TVIF_CHILDREN != 0 {
				nmtvdi.Item.CChildren = int32(item.ChildCount())
			}

		case win.TVN_ITEMEXPANDING:
			nmtv := (*win.NMTREEVIEW)(unsafe.Pointer(lParam))
			item := tv.handle2Item[nmtv.ItemNew.HItem]

			if nmtv.Action == win.TVE_EXPAND && tv.lazyPopulation {
				info := tv.item2Info[item]
				if len(info.child2Handle) == 0 {
					tv.insertChildren(item)
				}
			}

		case win.TVN_ITEMEXPANDED:
			nmtv := (*win.NMTREEVIEW)(unsafe.Pointer(lParam))
			item := tv.handle2Item[nmtv.ItemNew.HItem]

			switch nmtv.Action {
			case win.TVE_COLLAPSE:
				tv.expandedChangedPublisher.Publish(item)

			case win.TVE_COLLAPSERESET:

			case win.TVE_EXPAND:
				tv.expandedChangedPublisher.Publish(item)

			case win.TVE_EXPANDPARTIAL:

			case win.TVE_TOGGLE:
			}

		case win.NM_DBLCLK:
			tv.itemActivatedPublisher.Publish()

		case win.TVN_KEYDOWN:
			nmtvkd := (*win.NMTVKEYDOWN)(unsafe.Pointer(lParam))
			if nmtvkd.WVKey == uint16(KeyReturn) {
				tv.itemActivatedPublisher.Publish()
			}

		case win.TVN_SELCHANGED:
			nmtv := (*win.NMTREEVIEW)(unsafe.Pointer(lParam))

			tv.currItem = tv.handle2Item[nmtv.ItemNew.HItem]

			tv.currentItemChangedPublisher.Publish()
		}
	}

	return tv.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
