// Copyright 2015 andeya Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package surfer provides a high-concurrency web downloader written in Go.
// It supports GET/POST/HEAD methods and http/https, fixed UserAgent with cookie
// persistence or random UserAgents without cookies, and simulates browser behavior for login flows.
package surfer

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"sync"

	"github.com/andeya/gust/result"
)

var (
	surf         Surfer
	phantom      Surfer
	once_surf    sync.Once
	once_phantom sync.Once
	tempJsDir    = "./tmp"
	// phantomjsFile = filepath.Clean(path.Join(os.Getenv("GOPATH"), `/src/github.com/andeya/surfer/phantomjs/phantomjs`))
	phantomjsFile = `./phantomjs`
	cookieJar, _  = cookiejar.New(nil) // nil options never returns error
)

func Download(req Request) result.Result[*http.Response] {
	switch req.GetDownloaderID() {
	case SurfID:
		once_surf.Do(func() { surf = New(cookieJar) })
		return surf.Download(req)
	case PhantomJsID:
		once_phantom.Do(func() { phantom = NewPhantom(phantomjsFile, tempJsDir, cookieJar) })
		return phantom.Download(req)
	}
	return result.TryErr[*http.Response](errors.New("unknown downloader id"))
}

// DestroyJsFiles removes PhantomJS temporary JS files.
func DestroyJsFiles() {
	if pt, ok := phantom.(*Phantom); ok {
		pt.DestroyJsFiles()
	}
}

// Downloader represents an core of HTTP web browser for crawler.
type Surfer interface {
	// GET @param url string, header http.Header, cookies []*http.Cookie
	// HEAD @param url string, header http.Header, cookies []*http.Cookie
	// POST PostForm @param url, referer string, values url.Values, header http.Header, cookies []*http.Cookie
	// POST-M PostMultipart @param url, referer string, values url.Values, header http.Header, cookies []*http.Cookie
	Download(Request) result.Result[*http.Response]
}
