package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	var (
		conn 	net.Conn
		connUdp 	net.Conn
	)
	conn, err := net.Dial("tcp", proxyServer)
	if err != nil {
		fmt.Println(err)
	}
	proxyAddr, err := sock5Auth(conn, proxyServer)
	if err != nil {
		log.Print(err)
	}
	//需要提前创建udp连接，只能开一个socket ，不能多次创建socket,可以保证一次socket完成多次udp传送
	if isProxy {
		connUdp, err = net.Dial("udp", proxyAddr)
	}else{
		connUdp, err = net.Dial("udp", dnsServer)
	}
	if err != nil {
		fmt.Println(err.Error())
	}
	msg, n, t, err := sendUdp(connUdp, domain, dnsServer, isProxy)
	fmt.Println(msg, n, t)
	//pressTest(isProxy, proxyServer, dnsServer, domain) //压测入口函数
}
