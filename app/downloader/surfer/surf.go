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
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"crypto/tls"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/andeya/gust/option"
	"github.com/andeya/gust/result"
	"github.com/andeya/gust/syncutil"
	"github.com/andeya/pholcus/app/downloader/surfer/agent"
)

// Surf is the default Download implementation.
type Surf struct {
	CookieJar *cookiejar.Jar
}

// New creates a Surf downloader instance.
func New(jar ...*cookiejar.Jar) Surfer {
	s := new(Surf)
	if len(jar) != 0 {
		s.CookieJar = jar[0]
	} else {
		s.CookieJar, _ = cookiejar.New(nil) // nil options never returns error
	}
	return s
}

// Download implements the Surfer interface.
func (s *Surf) Download(req Request) (r result.Result[*http.Response]) {
	defer r.Catch()
	param := NewParam(req).Unwrap()
	param.header.Set("Connection", "close")
	param.client = s.buildClient(param)
	resp, err := s.httpRequest(param)
	result.RetVoid(err).Unwrap()

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		gzipReader, err := gzip.NewReader(resp.Body)
		result.RetVoid(err).Unwrap()
		resp.Body = gzipReader

	case "deflate":
		resp.Body = flate.NewReader(resp.Body)

	case "zlib":
		readCloser, err := zlib.NewReader(resp.Body)
		result.RetVoid(err).Unwrap()
		resp.Body = readCloser
	}

	resp = param.writeback(resp)

	return result.Ok(resp)
}

var dnsCache = &DnsCache{}

// DnsCache DNS cache
type DnsCache struct {
	ipPortLib syncutil.SyncMap[string, string]
}

// Reg registers ipPort to DNS cache.
func (d *DnsCache) Reg(addr, ipPort string) {
	d.ipPortLib.Store(addr, ipPort)
}

// Del deletes ipPort from DNS cache.
func (d *DnsCache) Del(addr string) {
	d.ipPortLib.Delete(addr)
}

// Query queries ipPort from DNS cache.
func (d *DnsCache) Query(addr string) option.Option[string] {
	return d.ipPortLib.Load(addr)
}

// buildClient creates, configures, and returns a *http.Client type.
func (s *Surf) buildClient(param *Param) *http.Client {
	client := &http.Client{
		CheckRedirect: param.checkRedirect,
	}

	if param.enableCookie {
		client.Jar = s.CookieJar
	}

	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			var (
				c     net.Conn
				err   error
				ipOpt = dnsCache.Query(addr)
			)
			ipPort := addr
			if ipOpt.IsSome() {
				ipPort = ipOpt.Unwrap()
				defer func() {
					if err != nil {
						dnsCache.Del(addr)
					}
				}()
			} else {
				defer func() {
					if err == nil {
						dnsCache.Reg(addr, c.RemoteAddr().String())
					}
				}()
			}
			c, err = net.DialTimeout(network, ipPort, param.dialTimeout)
			if err != nil {
				return nil, err
			}
			if param.connTimeout > 0 {
				c.SetDeadline(time.Now().Add(param.connTimeout))
			}
			return c, nil
		},
	}

	if param.proxy != nil {
		transport.Proxy = http.ProxyURL(param.proxy)
	}

	if strings.ToLower(param.url.Scheme) == "https" {
		transport.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
		transport.DisableCompression = true
	}
	client.Transport = transport
	return client
}

// send uses the given *http.Request to make an HTTP request.
func (s *Surf) httpRequest(param *Param) (resp *http.Response, err error) {
	req, err := http.NewRequest(param.method, param.url.String(), param.body)
	if err != nil {
		return nil, err
	}

	req.Header = param.header

	if param.tryTimes <= 0 {
		for {
			resp, err = param.client.Do(req)
			if err != nil {
				if !param.enableCookie {
					l := len(agent.UserAgents["common"])
					r := rand.New(rand.NewSource(time.Now().UnixNano()))
					req.Header.Set("User-Agent", agent.UserAgents["common"][r.Intn(l)])
				}
				time.Sleep(param.retryPause)
				continue
			}
			break
		}
	} else {
		for i := 0; i < param.tryTimes; i++ {
			resp, err = param.client.Do(req)
			if err != nil {
				if !param.enableCookie {
					l := len(agent.UserAgents["common"])
					r := rand.New(rand.NewSource(time.Now().UnixNano()))
					req.Header.Set("User-Agent", agent.UserAgents["common"][r.Intn(l)])
				}
				time.Sleep(param.retryPause)
				continue
			}
			break
		}
	}

	return resp, err
}
