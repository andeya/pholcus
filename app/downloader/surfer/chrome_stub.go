//go:build cover

package surfer

import (
	"errors"
	"net/http"
	"net/http/cookiejar"

	"github.com/andeya/gust/result"
)

type ChromeStub struct {
	CookieJar *cookiejar.Jar
}

func NewChrome(jar ...*cookiejar.Jar) Surfer {
	c := &ChromeStub{}
	if len(jar) != 0 {
		c.CookieJar = jar[0]
	} else {
		c.CookieJar, _ = cookiejar.New(nil)
	}
	return c
}

func (c *ChromeStub) Download(req Request) result.Result[*http.Response] {
	return result.TryErr[*http.Response](errors.New("chrome not available in coverage mode"))
}
