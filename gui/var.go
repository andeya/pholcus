package gui

import (
	. "github.com/henrylee2cn/pholcus/gui/model"
	"github.com/henrylee2cn/pholcus/spider"
	"github.com/lxn/walk"
)

var (
	toggleRunBtn *walk.PushButton
	setting      *walk.Composite
	mw           *walk.MainWindow
	runMode      *walk.GroupBox
	db           *walk.DataBinder
	ep           walk.ErrorPresenter
	mode         *walk.GroupBox
	host         *walk.Splitter
	spiderMenu   = NewSpiderMenu(spider.Menu)
)
