//go:build cover

package surfer

import "testing"

func TestDownloadPhantomJsIDStub(t *testing.T) {
	req := &mockRequest{downloaderID: PhantomJsID}
	r := Download(req)
	if r.IsOk() {
		t.Error("Download with PhantomJsID expected error in coverage mode")
	}
}

func TestDownloadChromeIDStub(t *testing.T) {
	req := &mockRequest{downloaderID: ChromeID}
	r := Download(req)
	if r.IsOk() {
		t.Error("Download with ChromeID expected error in coverage mode")
	}
}

func TestDestroyJsFilesStub(t *testing.T) {
	req := &mockRequest{downloaderID: PhantomJsID}
	Download(req)
	DestroyJsFiles()
}
