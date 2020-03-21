package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	var (
		isProxy bool = true //是否通过代理进行dns 查询
		proxyServer string = "10.20.47.145:1080" //sock5代理服务器的ip地址
		dnsServer string = "8.8.8.8:53" //需要请求的dns server的ip地址
	)
	//通过tcp连接到sock5的udp代理与代理进行验证
	conn, err := net.Dial("tcp", proxyServer)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("获取%s的tcp连接成功...\n", proxyServer)
	reader := bufio.NewReader(conn)
	defer conn.Close()
	//协商版本sock5，method的长度为1，认证的方式为0表示不认证
	licenseReq := clientLicenseReq{5, 1, [255]byte{0}}
	sendLicenseReq(conn, licenseReq)
	getLicenseResp(reader)
	//sock版本为5，3表示udp转发，rsv保留字段为0，1表示ipv4地址
	//ip只有在多主机的有意义，暂不关心
	connReq := clientConnReq{5, 3, 0, 1, [4]byte{0, 0, 0, 0}, [2]byte{0, 0}}
	sendClientConnReq(conn, connReq)
	proxyAddr, err := getConnResp(reader)
	if err != nil {
		fmt.Println(err)
	}
	//需要提前创建udp连接，只能开一个sock，不能多次创建sock
	var connUdp net.Conn
	if isProxy {
		connUdp, err = net.Dial("udp", proxyAddr)
	}else{
		connUdp, err = net.Dial("udp", dnsServer)
	}
	if err != nil {
		fmt.Println(err.Error())
	}
	defer connUdp.Close()
	//在传输UDP数据时，由于通过代理，所以需要按照一定的格式进行包装，在需要传送的数据之前添加一个报头
	remsg, n, t := Send(connUdp, "www.taobao.com", "123.125.81.6", 53, isProxy)
	fmt.Println(remsg, n, t)
	remsg2, n2, t2 := Send(connUdp, "www.baidu.com", "123.125.81.6", 53, isProxy)
	fmt.Println(remsg2, n2, t2)
}
