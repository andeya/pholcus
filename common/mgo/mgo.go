package mgo

import (
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/andeya/gust/result"
	"github.com/andeya/gust/syncutil"
	"github.com/andeya/pholcus/common/pool"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
)

type MgoSrc struct {
	*mgo.Session
}

var (
	session *mgo.Session
	err     error
)

var lazyPool = syncutil.NewLazyValueWithFunc(func() result.Result[pool.Pool] {
	gcSeconds := time.Duration(config.Conf().Mgo.ConnGCSeconds) * time.Second
	p := pool.ClassicPool(
		config.Conf().Mgo.ConnCap,
		config.Conf().Mgo.ConnCap/5,
		func() (pool.Src, error) {
			if err != nil || session.Ping() != nil {
				if session != nil {
					session.Close()
				}
				Refresh()
			}
			return &MgoSrc{session.Clone()}, err
		},
		gcSeconds,
	)
	return result.Ok(p)
})

func getPool() pool.Pool {
	return lazyPool.TryGetValue().Unwrap()
}

func Refresh() {
	session, err = mgo.Dial(config.Conf().Mgo.ConnStr)
	if err != nil {
		logs.Log().Error("MongoDB: %v\n", err)
	} else if err = session.Ping(); err != nil {
		logs.Log().Error("MongoDB: %v\n", err)
	} else {
		session.SetPoolLimit(config.Conf().Mgo.ConnCap)
	}
}

// Usable reports whether the MongoDB session is usable.
func (ms *MgoSrc) Usable() bool {
	if ms.Session == nil || ms.Session.Ping() != nil {
		return false
	}
	return true
}

// Reset is called when the resource is returned to the pool.
func (*MgoSrc) Reset() {}

// Close closes the session when removed from the pool.
func (ms *MgoSrc) Close() {
	if ms.Session == nil {
		return
	}
	ms.Session.Close()
}

// Error returns the last MongoDB connection error.
func Error() error {
	return err
}

// Call executes fn with a pooled MongoDB connection.
func Call(fn func(pool.Src) error) result.VoidResult {
	return getPool().Call(fn)
}

// Close shuts down the connection pool.
func Close() {
	getPool().Close()
}

// Len returns the current resource count.
func Len() int {
	return getPool().Len()
}

// DatabaseNames returns all database names.
func DatabaseNames() result.Result[[]string] {
	var names []string
	r := getPool().Call(func(src pool.Src) error {
		var e error
		names, e = src.(*MgoSrc).DatabaseNames()
		return e
	})
	if r.IsErr() {
		return result.TryErr[[]string](r.UnwrapErr())
	}
	return result.Ok(names)
}

// CollectionNames returns collection names for the given database.
func CollectionNames(dbname string) result.Result[[]string] {
	var names []string
	r := getPool().Call(func(src pool.Src) error {
		var e error
		names, e = src.(*MgoSrc).DB(dbname).CollectionNames()
		return e
	})
	if r.IsErr() {
		return result.TryErr[[]string](r.UnwrapErr())
	}
	return result.Ok(names)
}
