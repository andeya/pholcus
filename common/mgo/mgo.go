package mgo

import (
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/pool"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
)

type MgoSrc struct {
	*mgo.Session
}

var (
	connGcSecond = time.Duration(config.MGO_CONN_GC_SECOND) * 1e9
	session      *mgo.Session
	err          error
	MgoPool      = pool.ClassicPool(
		config.MGO_CONN_CAP,
		config.MGO_CONN_CAP/5,
		func() (pool.Src, error) {
			if err != nil || session.Ping() != nil {
				if session != nil {
					session.Close()
				}
				Refresh()
			}
			return &MgoSrc{session.Clone()}, err
		},
		connGcSecond)
)

func Refresh() {
	session, err = mgo.Dial(config.MGO_CONN_STR)
	if err != nil {
		logs.Log.Error("MongoDB: %v\n", err)
	} else if err = session.Ping(); err != nil {
		logs.Log.Error("MongoDB: %v\n", err)
	} else {
		session.SetPoolLimit(config.MGO_CONN_CAP)
	}
}

// Usable reports whether the MongoDB session is usable.
func (self *MgoSrc) Usable() bool {
	if self.Session == nil || self.Session.Ping() != nil {
		return false
	}
	return true
}

// Reset is called when the resource is returned to the pool.
func (*MgoSrc) Reset() {}

// Close closes the session when removed from the pool.
func (self *MgoSrc) Close() {
	if self.Session == nil {
		return
	}
	self.Session.Close()
}

// Error returns the last MongoDB connection error.
func Error() error {
	return err
}

// Call executes fn with a pooled MongoDB connection.
func Call(fn func(pool.Src) error) result.VoidResult {
	return MgoPool.Call(fn)
}

// Close shuts down the connection pool.
func Close() {
	MgoPool.Close()
}

// Len returns the current resource count.
func Len() int {
	return MgoPool.Len()
}

// DatabaseNames returns all database names.
func DatabaseNames() result.Result[[]string] {
	var names []string
	r := MgoPool.Call(func(src pool.Src) error {
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
	r := MgoPool.Call(func(src pool.Src) error {
		var e error
		names, e = src.(*MgoSrc).DB(dbname).CollectionNames()
		return e
	})
	if r.IsErr() {
		return result.TryErr[[]string](r.UnwrapErr())
	}
	return result.Ok(names)
}
