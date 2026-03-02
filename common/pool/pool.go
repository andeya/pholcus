// Package pool provides a generic resource pool with dynamic growth and idle resource recycling.
package pool

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/andeya/gust/result"
)

type (
	// Pool is a resource pool with a maximum capacity.
	Pool interface {
		Call(func(Src) error) result.VoidResult
		Close()
		Len() int
	}
	// classic implements a classic resource pool.
	classic struct {
		srcs     chan Src      // resources (Src must be a pointer type)
		capacity int           // pool capacity
		maxIdle  int           // max idle resources
		len      int           // current resource count
		factory  Factory       // resource factory
		gctime   time.Duration // idle resource recycling interval
		closed   bool          // whether the pool is closed
		sync.RWMutex
	}
	// Src is the resource interface.
	Src interface {
		Usable() bool
		Reset()
		Close()
	}
	// Factory creates a new resource.
	Factory func() (Src, error)
)

const (
	GC_TIME = 60e9
)

var (
	closedError = errors.New("资源池已关闭")
)

// ClassicPool creates a classic resource pool with the given capacity and idle recycling.
func ClassicPool(capacity, maxIdle int, factory Factory, gctime ...time.Duration) Pool {
	if len(gctime) == 0 {
		gctime = append(gctime, GC_TIME)
	}
	pool := &classic{
		srcs:     make(chan Src, capacity),
		capacity: capacity,
		maxIdle:  maxIdle,
		factory:  factory,
		gctime:   gctime[0],
		closed:   false,
	}
	go pool.gc()
	return pool
}

// Call invokes the callback with a resource from the pool.
func (self *classic) Call(callback func(Src) error) result.VoidResult {
	var src Src
	for {
		self.RLock()
		if self.closed {
			self.RUnlock()
			return result.TryErrVoid(closedError)
		}
		select {
		case src = <-self.srcs:
			self.RUnlock()
			if !src.Usable() {
				self.del(src)
				continue
			}
		default:
			self.RUnlock()
			err := self.incAuto()
			if err != nil {
				return result.TryErrVoid(err)
			}
			runtime.Gosched()
			continue
		}
		break
	}
	defer func() {
		if p := recover(); p != nil {
			_ = fmt.Errorf("%v", p)
		}
		self.recover(src)
	}()
	return result.RetVoid(callback(src))
}

// Close destroys the pool and releases all resources.
func (self *classic) Close() {
	self.Lock()
	defer self.Unlock()
	if self.closed {
		return
	}
	self.closed = true
	for i := len(self.srcs); i >= 0; i-- {
		(<-self.srcs).Close()
	}
	close(self.srcs)
	self.len = 0
}

// Len returns the current number of resources in the pool.
func (self *classic) Len() int {
	self.RLock()
	defer self.RUnlock()
	return self.len
}

// gc runs the idle resource recycling goroutine.
func (self *classic) gc() {
	for !self.isClosed() {
		self.Lock()
		extra := len(self.srcs) - self.maxIdle
		if extra > 0 {
			self.len -= extra
			for ; extra > 0; extra-- {
				(<-self.srcs).Close()
			}
		}
		self.Unlock()
		time.Sleep(self.gctime)
	}
}

func (self *classic) incAuto() error {
	self.Lock()
	defer self.Unlock()
	if self.len >= self.capacity {
		return nil
	}
	src, err := self.factory()
	if err != nil {
		return err
	}
	self.srcs <- src
	self.len++
	return nil
}

func (self *classic) del(src Src) {
	src.Close()
	self.Lock()
	self.len--
	self.Unlock()
}

func (self *classic) recover(src Src) {
	self.RLock()
	defer self.RUnlock()
	if self.closed {
		return
	}
	src.Reset()
	self.srcs <- src
}

func (self *classic) isClosed() bool {
	self.RLock()
	defer self.RUnlock()
	return self.closed
}
