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

package surfer

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type (
	Request interface {
		// url
		GetURL() string
		// GET POST POST-M HEAD
		GetMethod() string
		// POST values
		GetPostData() string
		// http header
		GetHeader() http.Header
		// enable http cookies
		GetEnableCookie() bool
		// dial tcp: i/o timeout
		GetDialTimeout() time.Duration
		// WSARecv tcp: i/o timeout
		GetConnTimeout() time.Duration
		// the max times of download
		GetTryTimes() int
		// the pause time of retry
		GetRetryPause() time.Duration
		// the download ProxyHost
		GetProxy() string
		// max redirect times
		GetRedirectTimes() int
		// select Surf ro PhomtomJS
		GetDownloaderID() int
	}

	// DefaultRequest is the default Request implementation.
	DefaultRequest struct {
		URL          string      // required
		Method       string      // GET POST POST-M HEAD (default GET)
		Header       http.Header // http header
		EnableCookie bool        // set in Spider.EnableCookie
		// POST values
		PostData string
		// dial tcp: i/o timeout
		DialTimeout time.Duration
		// WSARecv tcp: i/o timeout
		ConnTimeout time.Duration
		// the max times of download
		TryTimes int
		// how long pause when retry
		RetryPause time.Duration
		// max redirect times
		// when RedirectTimes equal 0, redirect times is ∞
		// when RedirectTimes less than 0, redirect times is 0
		RedirectTimes int
		// the download ProxyHost
		Proxy string

		// DownloaderID: 0=Surf (high concurrency), 1=PhantomJS (strong anti-block, slow)
		DownloaderID int

		once sync.Once // ensures prepare is called only once
	}
)

const (
	SurfID      = 0 // Surf downloader identifier
	PhantomJsID = 1 // PhantomJS downloader identifier
	ChromeID    = 2 // Chromium headless browser downloader identifier
	// Deprecated: Use PhantomJsID instead.
	PhomtomJsID        = PhantomJsID
	DefaultMethod      = "GET"           // default request method
	DefaultDialTimeout = 2 * time.Minute // default server request timeout
	DefaultConnTimeout = 2 * time.Minute // default download timeout
	DefaultTryTimes    = 3               // default max download attempts
	DefaultRetryPause  = 2 * time.Second // default pause before retry
)

func (dr *DefaultRequest) prepare() {
	if dr.Method == "" {
		dr.Method = DefaultMethod
	}
	dr.Method = strings.ToUpper(dr.Method)

	if dr.Header == nil {
		dr.Header = make(http.Header)
	}

	if dr.DialTimeout < 0 {
		dr.DialTimeout = 0
	} else if dr.DialTimeout == 0 {
		dr.DialTimeout = DefaultDialTimeout
	}

	if dr.ConnTimeout < 0 {
		dr.ConnTimeout = 0
	} else if dr.ConnTimeout == 0 {
		dr.ConnTimeout = DefaultConnTimeout
	}

	if dr.TryTimes == 0 {
		dr.TryTimes = DefaultTryTimes
	}

	if dr.RetryPause <= 0 {
		dr.RetryPause = DefaultRetryPause
	}

	if dr.DownloaderID != PhantomJsID && dr.DownloaderID != ChromeID {
		dr.DownloaderID = SurfID
	}
}

// GetURL returns the request URL.
func (dr *DefaultRequest) GetURL() string {
	dr.once.Do(dr.prepare)
	return dr.URL
}

// GetMethod returns the HTTP method (e.g. GET, POST).
func (dr *DefaultRequest) GetMethod() string {
	dr.once.Do(dr.prepare)
	return dr.Method
}

// GetPostData returns the POST request body.
func (dr *DefaultRequest) GetPostData() string {
	dr.once.Do(dr.prepare)
	return dr.PostData
}

// GetHeader returns the HTTP request headers.
func (dr *DefaultRequest) GetHeader() http.Header {
	dr.once.Do(dr.prepare)
	return dr.Header
}

// GetEnableCookie returns whether cookies are enabled.
func (dr *DefaultRequest) GetEnableCookie() bool {
	dr.once.Do(dr.prepare)
	return dr.EnableCookie
}

// GetDialTimeout returns the dial timeout.
func (dr *DefaultRequest) GetDialTimeout() time.Duration {
	dr.once.Do(dr.prepare)
	return dr.DialTimeout
}

// GetConnTimeout returns the connection read/write timeout.
func (dr *DefaultRequest) GetConnTimeout() time.Duration {
	dr.once.Do(dr.prepare)
	return dr.ConnTimeout
}

// GetTryTimes returns the max retry count.
func (dr *DefaultRequest) GetTryTimes() int {
	dr.once.Do(dr.prepare)
	return dr.TryTimes
}

// GetRetryPause returns the retry pause duration.
func (dr *DefaultRequest) GetRetryPause() time.Duration {
	dr.once.Do(dr.prepare)
	return dr.RetryPause
}

// GetProxy returns the proxy address.
func (dr *DefaultRequest) GetProxy() string {
	dr.once.Do(dr.prepare)
	return dr.Proxy
}

// GetRedirectTimes returns the max redirect count.
func (dr *DefaultRequest) GetRedirectTimes() int {
	dr.once.Do(dr.prepare)
	return dr.RedirectTimes
}

// GetDownloaderID returns the downloader ID (0=Surf, 1=PhantomJS, 2=Chrome).
func (dr *DefaultRequest) GetDownloaderID() int {
	dr.once.Do(dr.prepare)
	return dr.DownloaderID
}
