// Package proxy 提供了代理 IP 池管理与在线筛选功能。
package proxy

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andeya/gust/option"
	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/app/downloader/surfer"
	"github.com/andeya/pholcus/common/ping"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
)

// Proxy manages a pool of proxy IPs with online filtering and per-host sorting.
type Proxy struct {
	ipRegexp           *regexp.Regexp
	proxyIPTypeRegexp  *regexp.Regexp
	proxyUrlTypeRegexp *regexp.Regexp
	allIps             map[string]string
	all                map[string]bool
	online             int32
	usable             map[string]*ProxyForHost
	ticker             *time.Ticker
	tickMinute         int64
	threadPool         chan bool
	surf               surfer.Surfer
	sync.Mutex
}

const (
	CONN_TIMEOUT = 4 //4s
	DAIL_TIMEOUT = 4 //4s
	TRY_TIMES    = 3
	// Max concurrency for IP speed testing
	MAX_THREAD_NUM = 1000
)

// New creates and starts a Proxy that loads and filters proxy IPs from config.
func New() *Proxy {
	p := &Proxy{
		ipRegexp:           regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+`),
		proxyIPTypeRegexp:  regexp.MustCompile(`https?://([\w]*:[\w]*@)?[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+:[0-9]+`),
		proxyUrlTypeRegexp: regexp.MustCompile(`((https?|ftp):\/\/)?(([^:\n\r]+):([^@\n\r]+)@)?((www\.)?([^/\n\r:]+)):?([0-9]{1,5})?\/?([^?\n\r]+)?\??([^#\n\r]*)?#?([^\n\r]*)`),
		allIps:             map[string]string{},
		all:                map[string]bool{},
		usable:             make(map[string]*ProxyForHost),
		threadPool:         make(chan bool, MAX_THREAD_NUM),
		surf:               surfer.New(),
	}
	go p.Update()
	return p
}

// Count returns the number of online proxy IPs.
func (p *Proxy) Count() int32 {
	return p.online
}

// SetSurfForTest injects a Surfer for testing.
func (p *Proxy) SetSurfForTest(s surfer.Surfer) {
	p.surf = s
}

// Update refreshes the proxy IP list.
func (p *Proxy) Update() result.VoidResult {
	f, err := os.Open(config.Conf().ProxyFile)
	if err != nil {
		return result.TryErrVoid(err)
	}
	b, _ := io.ReadAll(f)
	f.Close()

	proxysIPType := p.proxyIPTypeRegexp.FindAllString(string(b), -1)
	for _, proxy := range proxysIPType {
		p.allIps[proxy] = p.ipRegexp.FindString(proxy)
		p.all[proxy] = false
	}

	proxysUrlType := p.proxyUrlTypeRegexp.FindAllString(string(b), -1)
	for _, proxy := range proxysUrlType {
		gvalue := p.proxyUrlTypeRegexp.FindStringSubmatch(proxy)
		p.allIps[proxy] = gvalue[6]
		p.all[proxy] = false
	}

	log.Printf(" *     Read proxy IPs: %v\n", len(p.all))

	p.findOnline()
	return result.OkVoid()
}

// findOnline filters proxy IPs that are online.
func (p *Proxy) findOnline() *Proxy {
	log.Printf(" *     Filtering online proxy IPs...")
	p.online = 0
	for proxy := range p.all {
		p.threadPool <- true
		go func(proxy string) {
			alive := ping.Ping(p.allIps[proxy], CONN_TIMEOUT).IsOk()
			p.Lock()
			p.all[proxy] = alive
			p.Unlock()
			if alive {
				atomic.AddInt32(&p.online, 1)
			}
			<-p.threadPool
		}(proxy)
	}
	for len(p.threadPool) > 0 {
		time.Sleep(0.2e9)
	}
	p.online = atomic.LoadInt32(&p.online)
	log.Printf(" *     Online proxy IP filtering complete, total: %v\n", p.online)

	return p
}

// UpdateTicker updates the ticker.
func (p *Proxy) UpdateTicker(tickMinute int64) {
	p.tickMinute = tickMinute
	p.ticker = time.NewTicker(time.Duration(p.tickMinute) * time.Minute)
	for _, proxyForHost := range p.usable {
		proxyForHost.curIndex++
		proxyForHost.isEcho = true
	}
}

// GetOne returns an unused proxy IP for this cycle and its response time.
func (p *Proxy) GetOne(u string) option.Option[string] {
	if p.online == 0 {
		return option.None[string]()
	}
	u2, _ := url.Parse(u)
	if u2.Host == "" {
		logs.Log().Informational(" *     [%v] Failed to set proxy IP, invalid target URL\n", u)
		return option.None[string]()
	}
	var key = u2.Host
	if strings.Count(key, ".") > 1 {
		key = key[strings.Index(key, ".")+1:]
	}

	p.Lock()
	defer p.Unlock()

	var ok = true
	var proxyForHost = p.usable[key]

	select {
	case <-p.ticker.C:
		proxyForHost.curIndex++
		if proxyForHost.curIndex >= proxyForHost.Len() {
			_, ok = p.testAndSort(key, u2.Scheme+"://"+u2.Host)
		}
		proxyForHost.isEcho = true

	default:
		if proxyForHost == nil {
			p.usable[key] = &ProxyForHost{
				proxys:    []string{},
				timedelay: []time.Duration{},
				isEcho:    true,
			}
			proxyForHost, ok = p.testAndSort(key, u2.Scheme+"://"+u2.Host)
		} else if l := proxyForHost.Len(); l == 0 {
			ok = false
		} else if proxyForHost.curIndex >= l {
			_, ok = p.testAndSort(key, u2.Scheme+"://"+u2.Host)
			proxyForHost.isEcho = true
		}
	}
	if !ok {
		logs.Log().Informational(" *     [%v] Failed to set proxy IP, no available proxy IPs\n", key)
		return option.None[string]()
	}
	curProxy := proxyForHost.proxys[proxyForHost.curIndex]
	if proxyForHost.isEcho {
		logs.Log().Informational(" *     Set proxy IP to [%v](%v)\n",
			curProxy,
			proxyForHost.timedelay[proxyForHost.curIndex],
		)
		proxyForHost.isEcho = false
	}
	return option.Some(curProxy)
}

// testAndSort tests and sorts proxy IPs for the given host.
func (p *Proxy) testAndSort(key string, testHost string) (*ProxyForHost, bool) {
	logs.Log().Informational(" *     [%v] Testing and sorting proxy IPs...", key)
	proxyForHost := p.usable[key]
	proxyForHost.proxys = []string{}
	proxyForHost.timedelay = []time.Duration{}
	proxyForHost.curIndex = 0
	for proxy, online := range p.all {
		if !online {
			continue
		}
		p.threadPool <- true
		go func(proxy string) {
			alive, timedelay := p.findUsable(proxy, testHost)
			if alive {
				proxyForHost.Mutex.Lock()
				proxyForHost.proxys = append(proxyForHost.proxys, proxy)
				proxyForHost.timedelay = append(proxyForHost.timedelay, timedelay)
				proxyForHost.Mutex.Unlock()
			}
			<-p.threadPool
		}(proxy)
	}
	for len(p.threadPool) > 0 {
		time.Sleep(0.2e9)
	}
	if proxyForHost.Len() > 0 {
		sort.Sort(proxyForHost)
		logs.Log().Informational(" *     [%v] Testing and sorting proxy IPs complete, available: %v\n", key, proxyForHost.Len())
		return proxyForHost, true
	}
	logs.Log().Informational(" *     [%v] Testing and sorting proxy IPs complete, no available proxy IPs\n", key)
	return proxyForHost, false
}

// findUsable tests proxy IP availability.
func (p *Proxy) findUsable(proxy string, testHost string) (alive bool, timedelay time.Duration) {
	t0 := time.Now()
	req := &request.Request{
		URL:         testHost,
		Method:      "HEAD",
		Header:      make(http.Header),
		DialTimeout: time.Second * time.Duration(DAIL_TIMEOUT),
		ConnTimeout: time.Second * time.Duration(CONN_TIMEOUT),
		TryTimes:    TRY_TIMES,
	}
	req.SetProxy(proxy)
	r := p.surf.Download(req)
	if r.IsErr() {
		return false, 0
	}
	resp := r.Unwrap()
	if resp == nil || resp.StatusCode != http.StatusOK {
		return false, 0
	}
	return true, time.Since(t0)
}
