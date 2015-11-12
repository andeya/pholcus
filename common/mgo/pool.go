package mgo

import (
	"errors"
	"github.com/henrylee2cn/pholcus/common/pool"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	mgo "gopkg.in/mgo.v2"
)

type MgoSrc struct {
	*mgo.Session
}

var MgoPool = pool.NewPool(new(MgoSrc), config.MGO.MAX_CONNS, 10)

// 新建数据库连接
func (*MgoSrc) New() pool.Src {
	session, err := mgo.Dial(config.MGO.CONN_STR)
	if err != nil {
		logs.Log.Error("MongoDB： %v", err)
		return nil
	}
	return &MgoSrc{Session: session}
}

// 判断连接是否失效
func (self *MgoSrc) Expired() bool {
	if self.Session == nil || self.Session.Ping() != nil {
		return true
	}
	return false
}

// 自毁方法，在被资源池删除时调用
func (self *MgoSrc) Close() {
	if self.Session == nil {
		return
	}
	self.Session.Close()
}

func (*MgoSrc) Clean() {}

// 打开指定数据库的集合
func Open(database, collection string) (s *MgoSrc, c *mgo.Collection, err error) {
	s = MgoPool.GetOne().(*MgoSrc)
	if s == nil {
		return nil, nil, errors.New("链接数据库超时！")
	}
	db := s.DB(database)
	if db == nil {
		Close(s)
		return nil, nil, errors.New("数据库连接错误！")
	}
	return s, s.DB(database).C(collection), nil
}

// 释放连接
func Close(m ...pool.Src) {
	MgoPool.Free(m...)
}

// 获取所有数据
func DatabaseNames() (names []string, err error) {
	return MgoPool.GetOne().(*MgoSrc).DatabaseNames()
}

// 获取数据库集合列表
func CollectionNames(dbname string) (names []string, err error) {
	return MgoPool.GetOne().(*MgoSrc).DB(dbname).CollectionNames()
}
