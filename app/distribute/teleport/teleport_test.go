package teleport

import (
	"encoding/json"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"
)

func freePort(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

func TestNew(t *testing.T) {
	tp := New()
	if tp == nil {
		t.Fatal("New returned nil")
	}
	if tp.GetMode() != 0 {
		t.Errorf("GetMode = %d, want 0", tp.GetMode())
	}
	if tp.CountNodes() != 0 {
		t.Errorf("CountNodes = %d, want 0", tp.CountNodes())
	}
}

func TestTP_SetUID(t *testing.T) {
	tp := New().(*TP)
	tp.SetUID("mine")
	if tp.uid != "mine" {
		t.Errorf("uid = %q, want mine", tp.uid)
	}
	tp.SetUID("client", "server")
	if tp.tpClient.serverUID != "server" {
		t.Errorf("serverUID = %q, want server", tp.tpClient.serverUID)
	}
}

func TestTP_SetPackHeader(t *testing.T) {
	tp := New().(*TP)
	tp.SetPackHeader("custom")
	if tp.Protocol.header != "custom" {
		t.Errorf("header = %q, want custom", tp.Protocol.header)
	}
}

func TestTP_SetApiRChan(t *testing.T) {
	tp := New().(*TP)
	tp.SetApiRChan(100)
	if cap(tp.apiReadChan) != 100 {
		t.Errorf("apiReadChan cap = %d, want 100", cap(tp.apiReadChan))
	}
}

func TestTP_SetConnWChan(t *testing.T) {
	tp := New().(*TP)
	tp.SetConnWChan(512)
	if tp.connWChanCap != 512 {
		t.Errorf("connWChanCap = %d, want 512", tp.connWChanCap)
	}
}

func TestTP_SetConnBuffer(t *testing.T) {
	tp := New().(*TP)
	tp.SetConnBuffer(2048)
	if tp.connBufferLen != 2048 {
		t.Errorf("connBufferLen = %d, want 2048", tp.connBufferLen)
	}
}

func TestTP_SetTimeout(t *testing.T) {
	tp := New().(*TP)
	d := 5 * time.Second
	tp.SetTimeout(d)
	if tp.timeout != d {
		t.Errorf("timeout = %v, want %v", tp.timeout, d)
	}
}

func TestTP_SetAPI(t *testing.T) {
	tp := New().(*TP)
	api := API{"test": &identity{}}
	tp.SetAPI(api)
	if tp.api["test"] == nil {
		t.Error("SetAPI did not set handler")
	}
}

func TestTP_ServerClient_Pipe(t *testing.T) {
	port := freePort(t)
	portStr := ":" + port

	serverTP := New().(*TP)
	serverTP.SetUID("server").SetTimeout(100 * time.Millisecond)
	serverTP.api["echo"] = &echoHandle{}
	serverTP.Server(portStr)
	time.Sleep(50 * time.Millisecond)

	clientTP := New().(*TP)
	clientTP.SetUID("client1").SetTimeout(100 * time.Millisecond)
	clientTP.api["echo"] = &echoHandle{}
	clientTP.Client("127.0.0.1", portStr)
	time.Sleep(100 * time.Millisecond)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		clientTP.Request("hello", "echo", "", "server")
	}()
	time.Sleep(200 * time.Millisecond)
	serverTP.Close()
	clientTP.Close()
	wg.Wait()
}

type echoHandle struct{}

func (*echoHandle) Process(receive *NetData) *NetData {
	return ReturnData(receive.Body, receive.Operation, receive.From, receive.To)
}

func TestTP_CloseSpecificNode(t *testing.T) {
	port := freePort(t)
	portStr := ":" + port

	serverTP := New().(*TP)
	serverTP.SetUID("server").SetTimeout(100 * time.Millisecond)
	serverTP.api["echo"] = &echoHandle{}
	serverTP.Server(portStr)
	time.Sleep(50 * time.Millisecond)

	clientTP := New().(*TP)
	clientTP.SetUID("client1").SetTimeout(100 * time.Millisecond)
	clientTP.api["echo"] = &echoHandle{}
	clientTP.Client("127.0.0.1", portStr)
	time.Sleep(100 * time.Millisecond)

	serverTP.Close("client1")
	clientTP.Close("server")
}

func TestConnect_Close(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	_, conn := NewConnect(client, 64, 16)
	if err := conn.Close(); err != nil {
		t.Errorf("Close() = %v", err)
	}
}

func TestTP_CheckRightsReject(t *testing.T) {
	port := freePort(t)
	portStr := ":" + port

	serverTP := New().(*TP)
	serverTP.SetUID("server").SetTimeout(100 * time.Millisecond)
	serverTP.api["echo"] = &echoHandle{}
	serverTP.Server(portStr)
	time.Sleep(50 * time.Millisecond)

	conn, err := net.Dial("tcp", "127.0.0.1"+portStr)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	nd := &NetData{From: "evil", To: "wrongserver", Operation: IDENTITY, Body: nil}
	data, _ := json.Marshal(nd)
	p := NewProtocol(DEFAULT_PACK_HEADER)
	packed := p.Packet(data)
	conn.Write(packed)
	conn.Close()
	time.Sleep(100 * time.Millisecond)
	serverTP.Close()
}

func TestDebugPrint(t *testing.T) {
	Debug = true
	defer func() { Debug = false }()
	debugPrintf("test %v", 1)
	debugPrintln("test")
}

func TestTP_GetConnAddr(t *testing.T) {
	tp := New().(*TP)
	if got := tp.getConnAddr("x"); got != "" {
		t.Errorf("getConnAddr(\"x\") = %q, want empty", got)
	}
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	_, c := NewConnect(client, 64, 16)
	c.Usable = true
	tp.connPool["node1"] = c
	if got := tp.getConnAddr("node1"); got == "" {
		t.Error("getConnAddr(\"node1\") = empty")
	}
}
