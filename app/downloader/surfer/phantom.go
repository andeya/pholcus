package surfer

import (
	"github.com/henrylee2cn/surfer/util"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 基于Phantomjs的下载器实现，作为surfer的补充
// 效率较surfer会慢很多，但是因为模拟浏览器，破防性更好
// 支持UserAgent/TryTimes/RetryPause/自定义js
type Phantom struct {
	FullPhantomjsName    string          //Phantomjs完整文件名
	FullTempJsFiles      map[string]bool //js临时文件存放完整文件名
	FullTempJsFilePrefix string          //js临时文件存放完整文件名前缀
	sync.Mutex
}

func NewPhantom(fullPhantomjsName, fullTempJsFilePrefix string) Surfer {
	phantom := &Phantom{
		FullPhantomjsName:    fullPhantomjsName,
		FullTempJsFilePrefix: fullTempJsFilePrefix,
		FullTempJsFiles:      map[string]bool{},
	}
	phantom.setFile(JS_CODE)
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

	jsfile, _ := self.setFile(JS_CODE)
	if js, ok := req.GetTemp("__JS__").(string); ok && js != "" {
		if _jsfile, err := self.setFile(js); err == nil {
			jsfile = _jsfile
		}
	}

	args := []string{jsfile, req.GetUrl(), encoding, strings.ToLower(param.header.Get("User-Agent"))}

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

	return
}

func (self *Phantom) setFile(js string) (string, error) {
	self.Mutex.Lock()
	defer self.Mutex.Unlock()
	jshash := util.MakeHash(js)
	fullFileName := self.FullTempJsFilePrefix + jshash
	if self.FullTempJsFiles[fullFileName] {
		return fullFileName, nil
	}
	if !filepath.IsAbs(fullFileName) {
		fullFileName, _ = filepath.Abs(fullFileName)
	}
	if !filepath.IsAbs(self.FullPhantomjsName) {
		self.FullPhantomjsName, _ = filepath.Abs(self.FullPhantomjsName)
	}

	// 创建/打开目录
	p, _ := filepath.Split(fullFileName)
	d, err := os.Stat(p)
	if err != nil || !d.IsDir() {
		if err := os.MkdirAll(p, 0777); err != nil {
			return "", err
		}
	}

	// 创建并写入文件
	f, _ := os.Create(fullFileName)
	f.Write([]byte(js))
	f.Close()
	self.FullTempJsFiles[fullFileName] = true
	return fullFileName, nil
}

const (
	JS_CODE = `//system 用于
	var system = require('system');
	var page = require('webpage').create();
	// console.log(system.args[0],system.args[1],system.args[2])
	page.settings.userAgent = 'Mozilla/5.0+(compatible;+Baiduspider/2.0;++http://www.baidu.com/search/spider.html)';
	if(system.args.length ==1){
		phantom.exit();
	}else{
		var url = system.args[1];
		var encode = system.args[2];

		if(encode != undefined){
			//设置编码
			phantom.outputEncoding=encode;
		}
		if(system.args[3] != undefined){
			
			//设置客户端代理设备
			page.settings.userAgent = system.args[3]
		}
		
		page.open(url, function (status) {
		    if (status !== 'success') {
		        console.log('Unable to access network');
		    } else {
		        // var ua = page.evaluate(function () {
		        //     return page.content;
		        // });
		        console.log(page.content);
		    }
		    phantom.exit();
		});
	}`
)
