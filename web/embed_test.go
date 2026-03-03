package web

import (
	"io/fs"
	"testing"
)

func TestViewsSubFS(t *testing.T) {
	sub := viewsSubFS()
	if sub == nil {
		t.Fatal("viewsSubFS() returned nil")
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		t.Errorf("viewsSubFS() index.html: %v", err)
	}
}
