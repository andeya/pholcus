package surfer

import (
	"github.com/henrylee2cn/surfer/util"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// 基于Phantomjs的下载器实现，作为surfer的补充
// 效率较surfer会慢很多，但是因为模拟浏览器，破防性更好
// 支持UserAgent/TryTimes/RetryPause/自定义js
type Phantom struct {
	FullPhantomjsName    string            //Phantomjs完整文件名
	FullTempJsFilePrefix string            //js临时文件存放完整文件名前缀
	jsFileMap            map[string]string //已存在的js文件
}

func NewPhantom(fullPhantomjsName, fullTempJsFilePrefix string) Surfer {
	phantom := &Phantom{
		FullPhantomjsName:    fullPhantomjsName,
		FullTempJsFilePrefix: fullTempJsFilePrefix,
		jsFileMap:            make(map[string]string),
	}
	phantom.createJsFile("get", getJs)
	phantom.createJsFile("post", postJs)
	return phantom
}

// 实现surfer下载器接口
func (self *Phantom) Download(req Request) (resp *http.Response, err error) {
	param, err := NewParam(req)
	if err != nil {
		return nil, err
	}
	resp = param.writeback(resp)

	encoding := strings.ToLower(param.header.Get("Content-Type"))
	if idx := strings.Index(encoding, "charset="); idx != -1 {
		encoding = strings.Trim(string(encoding[idx+8:]), ";")
		encoding = strings.Trim(encoding, " ")
	} else {
		encoding = "utf-8"
	}

	var args []string
	switch req.GetMethod() {
	case "GET":
		args = []string{
			self.jsFileMap["get"],
			req.GetUrl(),
			param.header.Get("Cookie"),
			encoding,
			param.header.Get("User-Agent"),
		}
	case "POST", "POST-M":
		args = []string{
			self.jsFileMap["post"],
			req.GetUrl(),
			param.header.Get("Cookie"),
			encoding,
			param.header.Get("User-Agent"),
			req.GetPostData().Encode(),
		}
	}

	for i := 0; i < param.tryTimes; i++ {
		cmd := exec.Command(self.FullPhantomjsName, args...)
		if resp.Body, err = cmd.StdoutPipe(); err != nil {
			time.Sleep(param.retryPause)
			continue
		}
		if cmd.Start() != nil || resp.Body == nil {
			time.Sleep(param.retryPause)
			continue
		}
		break
	}
	if err != nil {
		resp.Status = "200 OK"
		resp.StatusCode = 200
	}
	return
}

//销毁js临时文件
func (self *Phantom) DestroyJsFiles() {
	p, _ := path.Split(self.FullTempJsFilePrefix)
	if p == "" {
		return
	}
	for _, filename := range self.jsFileMap {
		os.Remove(filename)
	}
	if len(util.WalkFiles(p)) == 1 {
		os.Remove(p)
	}
}

func (self *Phantom) createJsFile(key, js string) {
	fullFileName := self.FullTempJsFilePrefix + "." + key
	if !filepath.IsAbs(fullFileName) {
		fullFileName, _ = filepath.Abs(fullFileName)
	}
	if !filepath.IsAbs(self.FullPhantomjsName) {
		self.FullPhantomjsName, _ = filepath.Abs(self.FullPhantomjsName)
	}

	// 创建/打开目录
	err := util.Mkdir(self.FullTempJsFilePrefix)
	if err != nil {
		return
	}

	// 创建并写入文件
	f, _ := os.Create(fullFileName)
	f.Write([]byte(js))
	f.Close()

	self.jsFileMap[key] = fullFileName
}

/*
* GET method
* system.args[0] == JSfile.js
* system.args[1] == url
* system.args[2] == cookie
* system.args[3] == pageEncode
* system.args[4] == userAgent
 */

const getJs string = `
	var system = require('system');
	var page = require('webpage').create();
	var url = system.args[1];
	var cookie = system.args[2];
	var pageEncode = system.args[3];
	var userAgent = system.args[4];
	page.onResourceRequested = function(requestData, request) {
		request.setHeader('Cookie', cookie)
	};
	phantom.outputEncoding=pageEncode;
	page.settings.userAgent = userAgent;
	page.open(url, function (status) {
		    if (status !== 'success') {
		        console.log('Unable to access network');
		    } else {
		        console.log(page.content);
		    }
		    phantom.exit();
	});
`

/*
* POST method
* system.args[0] == JSfile.js
* system.args[1] == url
* system.args[2] == cookie
* system.args[3] == pageEncode
* system.args[4] == userAgent
* system.args[5] == postdata
 */
const postJs string = `
	var system = require('system');
	var page = require('webpage').create();
	var url = system.args[1];
	var cookie = system.args[2];
	var pageEncode = system.args[3];
	var userAgent = system.args[4];
	var postdata = system.args[5];
	page.onResourceRequested = function(requestData, request) {
		request.setHeader('Cookie', cookie)
	};
	phantom.outputEncoding=pageEncode;
	page.settings.userAgent = userAgent;
	page.open(url, 'post', postdata, function (status) {
	    if (status !== 'success') {
	        console.log('Unable to access network');
	    } else {
	        console.log(page.content);
	    }
	    phantom.exit();
	});
`
