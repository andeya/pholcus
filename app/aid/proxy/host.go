package proxy

import (
	"sync"
	"time"
)

// ProxyForHost manages proxy IPs for a host, sorted by response time.
type ProxyForHost struct {
	curIndex  int // Index of current proxy IP
	proxys    []string
	timedelay []time.Duration
	isEcho    bool // Whether to print proxy switch info
	sync.Mutex
}

// Len implements sort.Interface.
func (ph *ProxyForHost) Len() int {
	return len(ph.proxys)
}

func (ph *ProxyForHost) Less(i, j int) bool {
	return ph.timedelay[i] < ph.timedelay[j]
}

func (ph *ProxyForHost) Swap(i, j int) {
	ph.proxys[i], ph.proxys[j] = ph.proxys[j], ph.proxys[i]
	ph.timedelay[i], ph.timedelay[j] = ph.timedelay[j], ph.timedelay[i]
}
