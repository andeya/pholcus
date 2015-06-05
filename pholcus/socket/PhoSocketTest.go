package main

import (
	// "bufio"
	"fmt"
	// "io/ioutil"
	"log"
	"net"
)

const (
	PhoSocketServer = "127.0.0.1:6010"
)

//建立连接

//func 接受

//func 发送

/*
*@服务端用
*
*PhoSoketLisent()为幽灵蛛socket监听函数
*PhoSocketServer为预定义常量：server:port
*输出类型为net.Listener,一个监听句柄
 */
func PhoSoketLisent() net.Listener {
	ln, err := net.Listen("tcp", PhoSocketServer)
	if err != nil {
		panic(err)
	}
	return ln
}

/*
*@客户端用
*
*PhoSoketDial()为幽灵蛛socket拨号函数，请求服务端
*PhoSocketServer为预定义常量：server:port
*输出类型为net.Conn,一个握手连接，下一步可以进行接收，发送
 */
func PhoSoketDial() net.Conn {
	conn, err := net.Dial("tcp", PhoSocketServer)
	if err != nil {
		panic(err)
	}
	return conn
}

/*
*@服务端用
*
*PhoSocketAccept()为幽灵蛛socket同意连接函数
*ln为一个监听句柄
*输出类型为net.Conn,一个握手连接，下一步可以进行接收，发送
 */
func PhoSocketAccept(ln net.Listener) net.Conn {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("get client connection error: ", err)
		}
		return conn
	}
}

/*
*@服务端用
*
*PhoSocketSendDataClose()为幽灵蛛socket数据发送函数，并且关闭连接
*conn为握手连接，sendData为要发送的数据
*通过conn给client发送Data
 */
func PhoSocketSendDataClose(conn net.Conn, sendData string) {
	// fmt.Fprintf(conn, sendData)
	conn.Write([]byte(sendData))
	conn.Close()
}

/*
*@共用
*
*PhoSocketSendData()为幽灵蛛socket数据发送函数,但是不关闭
*conn为握手连接，sendData为要发送的数据
*通过conn给client发送Data
 */
func PhoSocketSendData(conn net.Conn, sendData string) {
	// fmt.Fprintf(conn, sendData)
	conn.Write([]byte(sendData))
}

/*
*@共用
*
*PhoSocketAcceptData()为幽灵蛛socket数据接收函数
*conn为握手连接
*通过conn接收client发送来的Data
 */
func PhoSocketAcceptData(conn net.Conn) {
	// data, err := bufio.NewReader(conn).ReadString('\n')
	databuf := make([]byte, 4096)
	n, err := conn.Read(databuf)
	// data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Fatal("get client data error: ", err)
	}
	fmt.Printf("%#v\n", string(databuf[:n]))
}

//接收并发送，完关闭
func AcceptAndSendClose(conn net.Conn) {
	PhoSocketAcceptData(conn)
	PhoSocketSendDataClose(conn, "this is server\n")
}

//接收并发送，不关闭
func AcceptAndSend(conn net.Conn) {
	PhoSocketAcceptData(conn)
	PhoSocketSendData(conn, "this is server\n")
}
func main() {
	conn := PhoSoketDial()
	PhoSocketSendData(conn, "hello server\n")
	PhoSocketAcceptData(conn)

}
