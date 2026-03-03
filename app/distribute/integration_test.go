package distribute

import (
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/andeya/pholcus/app/distribute/teleport"
)

func freePort(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

func TestTP_ServerClient_Request(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	port := freePort(t)
	portStr := ":" + port

	tj := NewTaskJar()
	tj.Push(&Task{ID: 0, Limit: 100})
	serverTP := teleport.New().SetUID("server").SetAPI(MasterAPI(tj)).SetTimeout(100 * time.Millisecond)
	serverTP.Server(portStr)
	time.Sleep(50 * time.Millisecond)

	clientTP := teleport.New().SetUID("client").SetAPI(SlaveAPI(NewTaskJar())).SetTimeout(100 * time.Millisecond)
	clientTP.Client("127.0.0.1", portStr)
	time.Sleep(100 * time.Millisecond)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		clientTP.Request("", "task", "", "")
	}()
	time.Sleep(200 * time.Millisecond)
	serverTP.Close()
	clientTP.Close()
	wg.Wait()
}
