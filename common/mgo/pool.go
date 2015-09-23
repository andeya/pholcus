package mgo

import (
	"errors"

	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pool"
	mgo "gopkg.in/mgo.v2"
)

type MgoFish struct {
	*mgo.Session
}

var MgoPool = pool.NewPool(new(MgoFish), 1024)

// 新建数据库连接
func (*MgoFish) New() pool.Fish {
	session, err := mgo.Dial(config.MGO_OUTPUT.Host)
	if err != nil {
		logs.Log.Error("%v", err)
	}
	return &MgoFish{Session: session}
}

// 判断连接有效性
func (self *MgoFish) Usable() bool {
	if self.Session.Ping() != nil {
		return false
	}
	return true
}

// 自毁方法，在被资源池删除时调用
func (self *MgoFish) Close() {
	self.Session.Close()
}

func (*MgoFish) Clean() {}

// 打开指定数据库的集合
func Open(database, collection string) (s *MgoFish, c *mgo.Collection, err error) {
	s = MgoPool.GetOne().(*MgoFish)
	db := s.DB(database)
	if db == nil {
		Close(s)
		return nil, nil, errors.New("数据库连接错误！")
	}
	return s, s.DB(database).C(collection), nil
}

// 释放连接
func Close(m ...pool.Fish) {
	MgoPool.Free(m...)
}

// 获取所有数据
func DatabaseNames() (names []string, err error) {
	return MgoPool.GetOne().(*MgoFish).DatabaseNames()
}

// 获取数据库集合列表
func CollectionNames(dbname string) (names []string, err error) {
	return MgoPool.GetOne().(*MgoFish).DB(dbname).CollectionNames()
}
