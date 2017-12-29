// Copyright 2015 henrylee2cn Author. All Rights Reserved.
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
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type (
	// Phantom 基于Phantomjs的下载器实现，作为surfer的补充
	// 效率较surfer会慢很多，但是因为模拟浏览器，破防性更好
	// 支持UserAgent/TryTimes/RetryPause/自定义js
	Phantom struct {
		PhantomjsFile string            //Phantomjs完整文件名
		TempJsDir     string            //临时js存放目录
		jsFileMap     map[string]string //已存在的js文件
		CookieJar     *cookiejar.Jar
	}
	// Response 用于解析Phantomjs的响应内容
	Response struct {
		Cookies []string
		Body    string
		Error   string
		Header  []struct {
			Name  string
			Value string
		}
	}

	//给phantomjs传输cookie用
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
		phantom.CookieJar, _ = cookiejar.New(nil)
	}
	if !filepath.IsAbs(phantom.PhantomjsFile) {
		phantom.PhantomjsFile, _ = filepath.Abs(phantom.PhantomjsFile)
	}
	if !filepath.IsAbs(phantom.TempJsDir) {
		phantom.TempJsDir, _ = filepath.Abs(phantom.TempJsDir)
	}
	// 创建/打开目录
	err := os.MkdirAll(phantom.TempJsDir, 0777)
	if err != nil {
		log.Printf("[E] Surfer: %v\n", err)
		return phantom
	}
	phantom.createJsFile("js", js)
	return phantom
}

// 实现surfer下载器接口
func (self *Phantom) Download(req Request) (resp *http.Response, err error) {
	var encoding = "utf-8"
	if _, params, err := mime.ParseMediaType(req.GetHeader().Get("Content-Type")); err == nil {
		if cs, ok := params["charset"]; ok {
			encoding = strings.ToLower(strings.TrimSpace(cs))
		}
	}

	req.GetHeader().Del("Content-Type")

	param, err := NewParam(req)
	if err != nil {
		return nil, err
	}

	cookie := ""
	if req.GetEnableCookie() {
		httpCookies := self.CookieJar.Cookies(param.url)
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

	resp = param.writeback(resp)
	resp.Request.URL = param.url

	var args = []string{
		self.jsFileMap["js"],
		req.GetUrl(),
		cookie,
		encoding,
		param.header.Get("User-Agent"),
		req.GetPostData(),
		strings.ToLower(param.method),
		fmt.Sprint(int(req.GetDialTimeout() / time.Millisecond)),
	}

	for i := 0; i < param.tryTimes; i++ {
		if i != 0 {
			time.Sleep(param.retryPause)
		}

		cmd := exec.Command(self.PhantomjsFile, args...)
		if resp.Body, err = cmd.StdoutPipe(); err != nil {
			continue
		}
		err = cmd.Start()
		if err != nil || resp.Body == nil {
			continue
		}
		var b []byte
		b, err = ioutil.ReadAll(resp.Body)
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

		//设置header
		for _, h := range retResp.Header {
			resp.Header.Add(h.Name, h.Value)
		}

		//设置cookie
		for _, c := range retResp.Cookies {
			resp.Header.Add("Set-Cookie", c)
		}
		if req.GetEnableCookie() {
			if rc := resp.Cookies(); len(rc) > 0 {
				self.CookieJar.SetCookies(param.url, rc)
			}
		}
		resp.Body = ioutil.NopCloser(strings.NewReader(retResp.Body))
		break
	}

	if err == nil {
		resp.StatusCode = http.StatusOK
		resp.Status = http.StatusText(http.StatusOK)
	} else {
		resp.StatusCode = http.StatusBadGateway
		resp.Status = err.Error()
	}
	return
}

//销毁js临时文件
func (self *Phantom) DestroyJsFiles() {
	p, _ := filepath.Split(self.TempJsDir)
	if p == "" {
		return
	}
	for _, filename := range self.jsFileMap {
		os.Remove(filename)
	}
	if len(WalkDir(p)) == 1 {
		os.Remove(p)
	}
}

func (self *Phantom) createJsFile(fileName, jsCode string) {
	fullFileName := filepath.Join(self.TempJsDir, fileName)
	// 创建并写入文件
	f, _ := os.Create(fullFileName)
	f.Write([]byte(jsCode))
	f.Close()
	self.jsFileMap[fileName] = fullFileName
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

//输出参数
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
        // for (var j in response.headers) {//用javascript的for/in循环遍历对象的属性
        //     // var m = sprintf("AttrId[%d]Value[%d]", j, result.Attrs[j]);
        //     // message += m;
        //     // console.log(response.headers[j]);
        //     console.log(response.headers[j]["name"] + ":" + response.headers[j]["value"]);
        // }
        //
        // console.log("liguoqinjim received2------------------------------------------------");

        //在ret中加入header
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

    if (e.errorCode != 5) { //errorCode=5的情况和onResourceTimeout冲突
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
