// Package mgo provides MongoDB database connection and operation wrapper.
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

type (
	sessionProvider interface {
		DB(name string) dbProvider
		DatabaseNames() ([]string, error)
	}
	dbProvider interface {
		C(name string) collectionProvider
		CollectionNames() ([]string, error)
	}
	collectionProvider interface {
		Find(query interface{}) queryProvider
		Insert(docs ...interface{}) error
		Remove(selector interface{}) error
		Update(selector, update interface{}) error
		UpdateAll(selector, update interface{}) (*mgo.ChangeInfo, error)
		Upsert(selector, update interface{}) (*mgo.ChangeInfo, error)
	}
	queryProvider interface {
		Count() (int, error)
		Sort(fields ...string) queryProvider
		Skip(n int) queryProvider
		Limit(n int) queryProvider
		Select(selector interface{}) queryProvider
		All(result interface{}) error
	}
)

// MgoSrc wraps MongoDB session for connection pool.
type MgoSrc struct {
	*mgo.Session
}

type mgoSessionAdapter struct{ *MgoSrc }

func (m *mgoSessionAdapter) DB(name string) dbProvider {
	return &mgoDbAdapter{m.MgoSrc.DB(name)}
}

func (m *mgoSessionAdapter) DatabaseNames() ([]string, error) {
	return m.MgoSrc.DatabaseNames()
}

var getSessionFunc = func(src pool.Src) sessionProvider {
	return &mgoSessionAdapter{src.(*MgoSrc)}
}

type mgoDbAdapter struct{ *mgo.Database }

func (m *mgoDbAdapter) C(name string) collectionProvider {
	return &mgoCollectionAdapter{m.Database.C(name)}
}

func (m *mgoDbAdapter) CollectionNames() ([]string, error) {
	return m.Database.CollectionNames()
}

type mgoCollectionAdapter struct{ *mgo.Collection }

func (m *mgoCollectionAdapter) Find(query interface{}) queryProvider {
	return &mgoQueryAdapter{m.Collection.Find(query)}
}

func (m *mgoCollectionAdapter) Insert(docs ...interface{}) error {
	return m.Collection.Insert(docs...)
}

func (m *mgoCollectionAdapter) Remove(selector interface{}) error {
	return m.Collection.Remove(selector)
}

func (m *mgoCollectionAdapter) Update(selector, update interface{}) error {
	return m.Collection.Update(selector, update)
}

func (m *mgoCollectionAdapter) UpdateAll(selector, update interface{}) (*mgo.ChangeInfo, error) {
	return m.Collection.UpdateAll(selector, update)
}

func (m *mgoCollectionAdapter) Upsert(selector, update interface{}) (*mgo.ChangeInfo, error) {
	return m.Collection.Upsert(selector, update)
}

type mgoQueryAdapter struct{ *mgo.Query }

func (m *mgoQueryAdapter) Count() (int, error) {
	return m.Query.Count()
}

func (m *mgoQueryAdapter) Sort(fields ...string) queryProvider {
	return &mgoQueryAdapter{m.Query.Sort(fields...)}
}

func (m *mgoQueryAdapter) Skip(n int) queryProvider {
	return &mgoQueryAdapter{m.Query.Skip(n)}
}

func (m *mgoQueryAdapter) Limit(n int) queryProvider {
	return &mgoQueryAdapter{m.Query.Limit(n)}
}

func (m *mgoQueryAdapter) Select(selector interface{}) queryProvider {
	return &mgoQueryAdapter{m.Query.Select(selector)}
}

func (m *mgoQueryAdapter) All(result interface{}) error {
	return m.Query.All(result)
}

var (
	session *mgo.Session
	err     error
)

var testPool pool.Pool

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
	if testPool != nil {
		return testPool
	}
	return lazyPool.TryGetValue().Unwrap()
}

// Refresh re-establishes MongoDB connection.
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
		names, e = getSessionFunc(src).DatabaseNames()
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
		names, e = getSessionFunc(src).DB(dbname).CollectionNames()
		return e
	})
	if r.IsErr() {
		return result.TryErr[[]string](r.UnwrapErr())
	}
	return result.Ok(names)
}
