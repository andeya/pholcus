package beanstalkd

import (
	"net/url"
	"github.com/kr/beanstalk"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/pkg/errors"
)

type BeanstalkdClient struct {
	Conn     *beanstalk.Conn
	Tube string
}

func New() (*BeanstalkdClient, error) {
	tmp := new(BeanstalkdClient)
	host := config.BeanstalkdHost
	if host == "" {
		return nil, errors.New("beanstalk 主机为空")
	}
	tube := config.BeanstalkdTube
	if tube == "" {
		return nil, errors.New("tube name 为空")
	}
	conn, err := beanstalk.Dial("tcp", host)
	if err != nil {
		return nil, err
	}
	tmp.Tube = tube
	tmp.Conn = conn
	return tmp, nil
}

func (srv *BeanstalkdClient) Close() {
	if srv.Conn != nil {
		srv.Conn.Close()
	}
}

func (srv *BeanstalkdClient) Send(content url.Values) {
	if srv.Conn == nil {
		return
	}
	data := content.Encode()
	tube := &beanstalk.Tube{srv.Conn, srv.Tube}

	_, err := tube.Put([]byte(data), 1, 0, 0)
	if err != nil {
		logs.Log.Error("写入beanstalkd错误:%v,内容=%s", err, data)
		return
	}
	return
}
