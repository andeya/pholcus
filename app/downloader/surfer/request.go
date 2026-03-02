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
		GetUrl() string
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
		Url          string      // required
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
	// Deprecated: Use PhantomJsID instead.
	PhomtomJsID        = PhantomJsID
	DefaultMethod      = "GET"           // default request method
	DefaultDialTimeout = 2 * time.Minute // default server request timeout
	DefaultConnTimeout = 2 * time.Minute // default download timeout
	DefaultTryTimes    = 3               // default max download attempts
	DefaultRetryPause  = 2 * time.Second // default pause before retry
)

func (self *DefaultRequest) prepare() {
	if self.Method == "" {
		self.Method = DefaultMethod
	}
	self.Method = strings.ToUpper(self.Method)

	if self.Header == nil {
		self.Header = make(http.Header)
	}

	if self.DialTimeout < 0 {
		self.DialTimeout = 0
	} else if self.DialTimeout == 0 {
		self.DialTimeout = DefaultDialTimeout
	}

	if self.ConnTimeout < 0 {
		self.ConnTimeout = 0
	} else if self.ConnTimeout == 0 {
		self.ConnTimeout = DefaultConnTimeout
	}

	if self.TryTimes == 0 {
		self.TryTimes = DefaultTryTimes
	}

	if self.RetryPause <= 0 {
		self.RetryPause = DefaultRetryPause
	}

	if self.DownloaderID != PhantomJsID {
		self.DownloaderID = SurfID
	}
}

// url
func (self *DefaultRequest) GetUrl() string {
	self.once.Do(self.prepare)
	return self.Url
}

// GET POST POST-M HEAD
func (self *DefaultRequest) GetMethod() string {
	self.once.Do(self.prepare)
	return self.Method
}

// POST values
func (self *DefaultRequest) GetPostData() string {
	self.once.Do(self.prepare)
	return self.PostData
}

// http header
func (self *DefaultRequest) GetHeader() http.Header {
	self.once.Do(self.prepare)
	return self.Header
}

// enable http cookies
func (self *DefaultRequest) GetEnableCookie() bool {
	self.once.Do(self.prepare)
	return self.EnableCookie
}

// dial tcp: i/o timeout
func (self *DefaultRequest) GetDialTimeout() time.Duration {
	self.once.Do(self.prepare)
	return self.DialTimeout
}

// WSARecv tcp: i/o timeout
func (self *DefaultRequest) GetConnTimeout() time.Duration {
	self.once.Do(self.prepare)
	return self.ConnTimeout
}

// the max times of download
func (self *DefaultRequest) GetTryTimes() int {
	self.once.Do(self.prepare)
	return self.TryTimes
}

// the pause time of retry
func (self *DefaultRequest) GetRetryPause() time.Duration {
	self.once.Do(self.prepare)
	return self.RetryPause
}

// the download ProxyHost
func (self *DefaultRequest) GetProxy() string {
	self.once.Do(self.prepare)
	return self.Proxy
}

// max redirect times
func (self *DefaultRequest) GetRedirectTimes() int {
	self.once.Do(self.prepare)
	return self.RedirectTimes
}

// select Surf ro PhomtomJS
func (self *DefaultRequest) GetDownloaderID() int {
	self.once.Do(self.prepare)
	return self.DownloaderID
}
