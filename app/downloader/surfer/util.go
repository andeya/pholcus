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
	"io"

	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html/charset"
)

// AutoToUTF8 attempts to transcode response body to UTF-8 when using Surf.
// PhantomJS output is already UTF-8, so no transcoding is needed.
func AutoToUTF8(resp *http.Response) error {
	destReader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err == nil {
		resp.Body = &Body{
			ReadCloser: resp.Body,
			Reader:     destReader,
		}
	}
	return err
}

// BodyBytes reads the full response body.
func BodyBytes(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	return body, err
}

// URLEncode parses and encodes the URL, returning the result and any parse error.
func URLEncode(urlStr string) (*url.URL, error) {
	urlObj, err := url.Parse(urlStr)
	urlObj.RawQuery = urlObj.Query().Encode()
	return urlObj, err
}

// GetWDPath returns the working directory path (GOPATH).
func GetWDPath() string {
	wd := os.Getenv("GOPATH")
	if wd == "" {
		panic("GOPATH is not set in env.")
	}
	return wd
}

// IsDirExists checks whether the path is a directory.
func IsDirExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	}
	return fi.IsDir()
}

// IsFileExists checks whether the path is a file.
func IsFileExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	}
	return !fi.IsDir()
}

// WalkDir walks a directory, optionally filtering by suffix.
func WalkDir(targpath string, suffixes ...string) (dirlist []string) {
	if !filepath.IsAbs(targpath) {
		targpath, _ = filepath.Abs(targpath)
	}
	err := filepath.Walk(targpath, func(retpath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !f.IsDir() {
			return nil
		}
		if len(suffixes) == 0 {
			dirlist = append(dirlist, retpath)
			return nil
		}
		for _, suffix := range suffixes {
			if strings.HasSuffix(retpath, suffix) {
				dirlist = append(dirlist, retpath)
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("utils.WalkDir: %v\n", err)
		return
	}

	return
}

// ExtractHomepage returns the scheme + host portion of a URL, e.g.
// "https://www.baidu.com/s?wd=go" → "https://www.baidu.com".
func ExtractHomepage(rawURL string) string {
	idx := strings.Index(rawURL, "://")
	if idx < 0 {
		return ""
	}
	rest := rawURL[idx+3:]
	slash := strings.Index(rest, "/")
	if slash < 0 {
		return rawURL
	}
	return rawURL[:idx+3+slash]
}

// Body wraps Response.Body with a custom Reader for transcoding.
type Body struct {
	io.ReadCloser
	io.Reader
}

func (b *Body) Read(p []byte) (int, error) {
	return b.Reader.Read(p)
}
