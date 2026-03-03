// Package beanstalkd 提供了 Beanstalkd 任务队列的客户端封装。
package beanstalkd

import (
	"net/url"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
	"github.com/kr/beanstalk"
	"github.com/pkg/errors"
)

// BeanstalkdClient wraps a beanstalk connection and tube for job queuing.
type BeanstalkdClient struct {
	Conn *beanstalk.Conn
	Tube string
}

// New creates a new BeanstalkdClient using config.Conf().Beanstalkd.
func New() result.Result[*BeanstalkdClient] {
	return NewFromConfig(config.Conf().Beanstalkd)
}

// NewFromConfig creates a BeanstalkdClient from the given config.
func NewFromConfig(cfg config.BeanstalkdConfig) result.Result[*BeanstalkdClient] {
	tmp := new(BeanstalkdClient)
	if cfg.Host == "" {
		return result.TryErr[*BeanstalkdClient](errors.New("beanstalk host is empty"))
	}
	if cfg.Tube == "" {
		return result.TryErr[*BeanstalkdClient](errors.New("tube name is empty"))
	}
	conn, err := beanstalk.Dial("tcp", cfg.Host)
	if err != nil {
		return result.TryErr[*BeanstalkdClient](err)
	}
	tmp.Tube = cfg.Tube
	tmp.Conn = conn
	return result.Ok(tmp)
}

// Close closes the beanstalk connection.
func (srv *BeanstalkdClient) Close() {
	if srv.Conn != nil {
		srv.Conn.Close()
	}
}

// Send encodes content as URL values and puts it into the configured tube.
func (srv *BeanstalkdClient) Send(content url.Values) result.VoidResult {
	if srv.Conn == nil {
		return result.OkVoid()
	}
	data := content.Encode()
	tube := &beanstalk.Tube{Conn: srv.Conn, Name: srv.Tube}

	_, err := tube.Put([]byte(data), 1, 0, 0)
	if err != nil {
		logs.Log().Error("beanstalkd write error: %v, content=%s", err, data)
		return result.TryErrVoid(err)
	}
	return result.OkVoid()
}
