package surfer

import (
	"net/http"
	"net/url"
	"time"
)

type Request interface {
	GetUrl() string
	// GET POST POST-M HEAD
	GetMethod() string
	// POST values
	GetPostData() url.Values
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

	GetTemp(key string) interface{}
}
