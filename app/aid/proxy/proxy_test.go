package proxy

import (
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
)

// 可用于筛选有效IP
func TestUpdateSort(t *testing.T) {
	self := &Proxy{
		ipRegexp:    regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+`),
		proxyRegexp: regexp.MustCompile(`http[s]?://[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+:[0-9]+`),
		ipMap:       map[string]string{},
		usable:      map[string]bool{},
		pingPool:    make(chan bool, MAX_THREAD_NUM),
	}
	once.Do(mkdir)

	f, _ := os.OpenFile(config.PROXY_FULL_FILE_NAME, os.O_CREATE|os.O_RDWR, 0660)
	b, _ := ioutil.ReadAll(f)
	f.Close()

	proxys := self.proxyRegexp.FindAllString(string(b), -1)
	for _, proxy := range proxys {
		self.ipMap[proxy] = self.ipRegexp.FindString(proxy)
		self.usable[proxy] = true
		// fmt.Printf("+ 代理IP %v：%v\n", i, proxy)
	}
	logs.Log.Informational(" *     读取代理IP: %v 条\n", len(self.usable))

	self.updateSort()

	var a string
	for _, k := range self.speed {
		a += k + "\n"
	}

	os.Remove(config.PROXY_FULL_FILE_NAME)

	f2, _ := os.OpenFile(config.PROXY_FULL_FILE_NAME, os.O_CREATE|os.O_RDWR, 0660)
	f2.Write([]byte(a))
	f2.Close()
	t.Log("筛选有效IP完成")
}
