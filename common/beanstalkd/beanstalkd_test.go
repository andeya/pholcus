package beanstalkd

import (
	"net"
	"net/url"
	"testing"

	"github.com/andeya/pholcus/config"
	"github.com/kr/beanstalk"
)

func TestNewFromConfig_EmptyHost(t *testing.T) {
	r := NewFromConfig(config.BeanstalkdConfig{Host: "", Tube: "pholcus"})
	if r.IsOk() {
		t.Fatal("expected Err for empty host")
	}
	if err := r.UnwrapErr(); err == nil || err.Error() != "beanstalk host is empty" {
		t.Errorf("UnwrapErr() = %v, want beanstalk host is empty", err)
	}
}

func TestNewFromConfig_EmptyTube(t *testing.T) {
	r := NewFromConfig(config.BeanstalkdConfig{Host: "localhost:11300", Tube: ""})
	if r.IsOk() {
		t.Fatal("expected Err for empty tube")
	}
	if err := r.UnwrapErr(); err == nil || err.Error() != "tube name is empty" {
		t.Errorf("UnwrapErr() = %v, want tube name is empty", err)
	}
}

func TestNewFromConfig_ConnectionError(t *testing.T) {
	r := NewFromConfig(config.BeanstalkdConfig{Host: "127.0.0.1:1", Tube: "pholcus"})
	if r.IsOk() {
		t.Fatal("expected Err for connection failure")
	}
	if r.UnwrapErr() == nil {
		t.Error("UnwrapErr() should not be nil")
	}
}

func TestClose_NilConn(t *testing.T) {
	client := &BeanstalkdClient{Conn: nil, Tube: "pholcus"}
	client.Close()
}

func TestSend_NilConn(t *testing.T) {
	client := &BeanstalkdClient{Conn: nil, Tube: "pholcus"}
	r := client.Send(url.Values{"k": {"v"}})
	if r.IsErr() {
		t.Errorf("Send with nil Conn should return OkVoid, got Err: %v", r.UnwrapErr())
	}
}

func TestSend_EmptyValues(t *testing.T) {
	client := &BeanstalkdClient{Conn: nil, Tube: "pholcus"}
	r := client.Send(url.Values{})
	if r.IsErr() {
		t.Errorf("Send empty values should return OkVoid, got Err: %v", r.UnwrapErr())
	}
}

func TestClose_WithConn(t *testing.T) {
	c1, c2 := net.Pipe()
	_ = c2.Close()
	conn := beanstalk.NewConn(c1)
	client := &BeanstalkdClient{Conn: conn, Tube: "pholcus"}
	client.Close()
}

func TestSend_PutError(t *testing.T) {
	c1, c2 := net.Pipe()
	_ = c2.Close()
	conn := beanstalk.NewConn(c1)
	client := &BeanstalkdClient{Conn: conn, Tube: "pholcus"}
	defer client.Close()
	r := client.Send(url.Values{"k": {"v"}})
	if r.IsOk() {
		t.Error("Send with broken conn should return Err")
	}
}

func TestNew(t *testing.T) {
	c := config.Conf()
	origHost, origTube := c.Beanstalkd.Host, c.Beanstalkd.Tube
	defer func() {
		c.Beanstalkd.Host = origHost
		c.Beanstalkd.Tube = origTube
	}()
	c.Beanstalkd.Host = "127.0.0.1:1"
	c.Beanstalkd.Tube = "pholcus"
	r := New()
	if r.IsOk() {
		t.Fatal("New with invalid host should return Err")
	}
}
