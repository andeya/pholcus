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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/andeya/gust/result"
)

type (
	// Phantom is a PhantomJS-based downloader, complementing Surf.
	// Slower than Surf but better at bypassing anti-scraping due to browser simulation.
	// Supports UserAgent, TryTimes, RetryPause, and custom JS.
	Phantom struct {
		PhantomjsFile string            // full path to PhantomJS executable
		TempJsDir     string            // directory for temporary JS files
		jsFileMap     map[string]string // existing JS files
		CookieJar     *cookiejar.Jar
	}
	// Response parses PhantomJS response content.
	Response struct {
		Cookies []string
		Body    string
		Error   string
		Header  []struct {
			Name  string
			Value string
		}
	}

	// Cookie is used to pass cookies to PhantomJS.
	Cookie struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Domain string `json:"domain"`
		Path   string `json:"path"`
	}
)

func NewPhantom(phantomjsFile, tempJsDir string, jar ...*cookiejar.Jar) Surfer {
	phantom := &Phantom{
		PhantomjsFile: phantomjsFile,
		TempJsDir:     tempJsDir,
		jsFileMap:     make(map[string]string),
	}
	if len(jar) != 0 {
		phantom.CookieJar = jar[0]
	} else {
		phantom.CookieJar, _ = cookiejar.New(nil) // nil options never returns error
	}
	if !filepath.IsAbs(phantom.PhantomjsFile) {
		if absPath, err := filepath.Abs(phantom.PhantomjsFile); err != nil {
			log.Printf("[E] Surfer: filepath.Abs(%q): %v", phantom.PhantomjsFile, err)
		} else {
			phantom.PhantomjsFile = absPath
		}
	}
	if !filepath.IsAbs(phantom.TempJsDir) {
		if absPath, err := filepath.Abs(phantom.TempJsDir); err != nil {
			log.Printf("[E] Surfer: filepath.Abs(%q): %v", phantom.TempJsDir, err)
		} else {
			phantom.TempJsDir = absPath
		}
	}
	err := os.MkdirAll(phantom.TempJsDir, 0777)
	if err != nil {
		log.Printf("[E] Surfer: %v\n", err)
		return phantom
	}
	phantom.createJsFile("js", js)
	return phantom
}

// Download implements the Surfer interface.
func (p *Phantom) Download(req Request) (r result.Result[*http.Response]) {
	defer r.Catch()
	var encoding = "utf-8"
	if _, params, err := mime.ParseMediaType(req.GetHeader().Get("Content-Type")); err == nil {
		if cs, ok := params["charset"]; ok {
			encoding = strings.ToLower(strings.TrimSpace(cs))
		}
	}

	req.GetHeader().Del("Content-Type")

	param := NewParam(req).Unwrap()

	cookie := ""
	if req.GetEnableCookie() {
		httpCookies := p.CookieJar.Cookies(param.url)
		if len(httpCookies) > 0 {
			surferCookies := make([]*Cookie, len(httpCookies))

			for n, c := range httpCookies {
				surferCookie := &Cookie{Name: c.Name, Value: c.Value, Domain: param.url.Host, Path: "/"}
				surferCookies[n] = surferCookie
			}

			c, err := json.Marshal(surferCookies)
			if err != nil {
				log.Printf("cookie marshal error:%v", err)
			}
			cookie = string(c)
		}
	}

	resp := param.writeback(nil)
	resp.Request.URL = param.url

	var args = []string{
		p.jsFileMap["js"],
		req.GetURL(),
		cookie,
		encoding,
		param.header.Get("User-Agent"),
		req.GetPostData(),
		strings.ToLower(param.method),
		fmt.Sprint(int(req.GetDialTimeout() / time.Millisecond)),
	}
	if req.GetProxy() != "" {
		args = append([]string{"--proxy=" + req.GetProxy()}, args...)
	}

	var err error
	for i := 0; i < param.tryTimes; i++ {
		if i != 0 {
			time.Sleep(param.retryPause)
		}

		cmd := exec.Command(p.PhantomjsFile, args...)
		if resp.Body, err = cmd.StdoutPipe(); err != nil {
			continue
		}
		err = cmd.Start()
		if err != nil || resp.Body == nil {
			continue
		}
		var b []byte
		b, err = io.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		retResp := Response{}
		err = json.Unmarshal(b, &retResp)
		if err != nil {
			continue
		}

		if retResp.Error != "" {
			log.Printf("phantomjs response error:%s", retResp.Error)
			continue
		}

		for _, h := range retResp.Header {
			resp.Header.Add(h.Name, h.Value)
		}

		for _, c := range retResp.Cookies {
			resp.Header.Add("Set-Cookie", c)
		}
		if req.GetEnableCookie() {
			if rc := resp.Cookies(); len(rc) > 0 {
				p.CookieJar.SetCookies(param.url, rc)
			}
		}
		resp.Body = io.NopCloser(strings.NewReader(retResp.Body))
		err = nil
		break
	}

	if err == nil {
		resp.StatusCode = http.StatusOK
		resp.Status = http.StatusText(http.StatusOK)
	} else {
		resp.StatusCode = http.StatusBadGateway
		resp.Status = err.Error()
	}
	return result.Ok(resp)
}

// DestroyJsFiles removes temporary JS files.
func (p *Phantom) DestroyJsFiles() {
	dir, _ := filepath.Split(p.TempJsDir)
	if dir == "" {
		return
	}
	for _, filename := range p.jsFileMap {
		os.Remove(filename)
	}
	if len(WalkDir(dir)) == 1 {
		os.Remove(dir)
	}
}

func (p *Phantom) createJsFile(fileName, jsCode string) {
	fullFileName := filepath.Join(p.TempJsDir, fileName)
	f, _ := os.Create(fullFileName)
	f.Write([]byte(jsCode))
	f.Close()
	p.jsFileMap[fileName] = fullFileName
}

/*
* system.args[0] == js
* system.args[1] == url
* system.args[2] == cookie
* system.args[3] == pageEncode
* system.args[4] == userAgent
* system.args[5] == postdata
* system.args[6] == method
* system.args[7] == timeout
 */
const js string = `
var system = require('system');
var page = require('webpage').create();
var url = system.args[1];
var cookie = system.args[2];
var pageEncode = system.args[3];
var userAgent = system.args[4];
var postdata = system.args[5];
var method = system.args[6];
var timeout = system.args[7];

var ret = new Object();
var exit = function () {
    console.log(JSON.stringify(ret));
    phantom.exit();
};

// output params
// console.log("url=" + url);
// console.log("cookie=" + cookie);
// console.log("pageEncode=" + pageEncode);
// console.log("userAgent=" + userAgent);
// console.log("postdata=" + postdata);
// console.log("method=" + method);
// console.log("timeout=" + timeout);

// ret += (url + "\n");
// ret += (cookie + "\n");
// ret += (pageEncode + "\n");
// ret += (userAgent + "\n");
// ret += (postdata + "\n");
// ret += (method + "\n");
// ret += (timeout + "\n");
// exit();

phantom.outputEncoding = pageEncode;
page.settings.userAgent = userAgent;
page.settings.resourceTimeout = timeout;
page.settings.XSSAuditingEnabled = true;

function addCookie() {
    if (cookie != "") {
        var cookies = JSON.parse(cookie);
        for (var i = 0; i < cookies.length; i++) {
            var c = cookies[i];

            phantom.addCookie({
                'name': c.name, /* required property */
                'value': c.value, /* required property */
                'domain': c.domain,
                'path': c.path, /* required property */
            });
        }
    }
}

addCookie();

page.onResourceRequested = function (requestData, request) {

};
page.onResourceReceived = function (response) {
    if (response.stage === "end") {
        // console.log("liguoqinjim received1------------------------------------------------");
        // console.log("url=" + response.url);
        //
        // for (var j in response.headers) { // iterate object properties with for/in
        //     // var m = sprintf("AttrId[%d]Value[%d]", j, result.Attrs[j]);
        //     // message += m;
        //     // console.log(response.headers[j]);
        //     console.log(response.headers[j]["name"] + ":" + response.headers[j]["value"]);
        // }
        //
        // console.log("liguoqinjim received2------------------------------------------------");

        ret["Header"] = response.headers;
    }
};
page.onError = function (msg, trace) {
    ret["Error"] = msg;
    exit();
};
page.onResourceTimeout = function (e) {
    // console.log("phantomjs onResourceTimeout error");
    // console.log(e.errorCode);   // it'll probably be 408
    // console.log(e.errorString); // it'll probably be 'Network timeout on resource'
    // console.log(e.url);         // the url whose request timed out
    // phantom.exit(1);
    ret["Error"] = "onResourceTimeout";
    exit();
};
page.onResourceError = function (e) {
    // console.log("onResourceError");
    // console.log("1:" + e.errorCode + "," + e.errorString);

    if (e.errorCode != 5) { // errorCode=5 conflicts with onResourceTimeout
        ret["Error"] = "onResourceError";
        exit();
    }
};
page.onLoadFinished = function (status) {
    if (status !== 'success') {
        ret["Error"] = "status=" + status;
        exit();
    } else {
        var cookies = new Array();
        for (var i in page.cookies) {
            var cookie = page.cookies[i];
            var c = cookie["name"] + "=" + cookie["value"];
            for (var obj in cookie) {
                if (obj == 'name' || obj == 'value') {
                    continue;
                }
                if (obj == "httponly" || obj == "secure") {
                    if (cookie[obj] == true) {
                        c += ";" + obj;
                    }
                } else {
                    c += "; " + obj + "=" + cookie[obj];
                }
            }
            cookies[i] = c;
        }
        if (page.content.indexOf("body") != -1) {
            ret["Cookies"] = cookies;
            ret["Body"] = page.content;

            // ret = JSON.stringify(resp);
            exit();
        }
    }
};

page.open(url, method, postdata, function (status) {
});
`
