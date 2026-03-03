// Package web provides HTTP service, routing, and embedded resources for the Web interface.
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
