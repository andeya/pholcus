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
func (c *classic) Call(callback func(Src) error) result.VoidResult {
	var src Src
	for {
		c.RLock()
		if c.closed {
			c.RUnlock()
			return result.TryErrVoid(closedError)
		}
		select {
		case src = <-c.srcs:
			c.RUnlock()
			if !src.Usable() {
				c.del(src)
				continue
			}
		default:
			c.RUnlock()
			err := c.incAuto()
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
		c.recover(src)
	}()
	return result.RetVoid(callback(src))
}

// Close destroys the pool and releases all resources.
func (c *classic) Close() {
	c.Lock()
	defer c.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	for i := len(c.srcs); i >= 0; i-- {
		(<-c.srcs).Close()
	}
	close(c.srcs)
	c.len = 0
}

// Len returns the current number of resources in the pool.
func (c *classic) Len() int {
	c.RLock()
	defer c.RUnlock()
	return c.len
}

// gc runs the idle resource recycling goroutine.
func (c *classic) gc() {
	for !c.isClosed() {
		c.Lock()
		extra := len(c.srcs) - c.maxIdle
		if extra > 0 {
			c.len -= extra
			for ; extra > 0; extra-- {
				(<-c.srcs).Close()
			}
		}
		c.Unlock()
		time.Sleep(c.gctime)
	}
}

func (c *classic) incAuto() error {
	c.Lock()
	defer c.Unlock()
	if c.len >= c.capacity {
		return nil
	}
	src, err := c.factory()
	if err != nil {
		return err
	}
	c.srcs <- src
	c.len++
	return nil
}

func (c *classic) del(src Src) {
	src.Close()
	c.Lock()
	c.len--
	c.Unlock()
}

func (c *classic) recover(src Src) {
	c.RLock()
	defer c.RUnlock()
	if c.closed {
		return
	}
	src.Reset()
	c.srcs <- src
}

func (c *classic) isClosed() bool {
	c.RLock()
	defer c.RUnlock()
	return c.closed
}
