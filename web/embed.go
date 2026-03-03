// Package web 提供了 Web 界面的 HTTP 服务、路由和嵌入式资源。
package web

import (
	"embed"
	"io/fs"
)

//go:embed views
var viewsFS embed.FS

// viewsSubFS returns a sub-filesystem rooted at the "views" directory.
func viewsSubFS() fs.FS {
	sub, err := fs.Sub(viewsFS, "views")
	if err != nil {
		panic(err)
	}
	return sub
}
