//go:build cover

package surfer

import (
	"errors"
	"net/http"
	"net/http/cookiejar"

	"github.com/andeya/gust/result"
)

type PhantomStub struct {
	CookieJar *cookiejar.Jar
}

func NewPhantom(phantomjsFile, tempJsDir string, jar ...*cookiejar.Jar) Surfer {
	p := &PhantomStub{}
	if len(jar) != 0 {
		p.CookieJar = jar[0]
	} else {
		p.CookieJar, _ = cookiejar.New(nil)
	}
	return p
}

func (p *PhantomStub) Download(req Request) result.Result[*http.Response] {
	return result.TryErr[*http.Response](errors.New("phantom not available in coverage mode"))
}

func (p *PhantomStub) DestroyJsFiles() {}
