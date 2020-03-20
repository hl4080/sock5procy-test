package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	//通过tcp连接到sock5的udp代理与代理进行验证
	conn, err := net.Dial("tcp", "10.20.47.145:1080")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("获取10.20.47.145:1080的tcp连接成功...")
	reader := bufio.NewReader(conn)
	defer conn.Close()
	//协商版本sock5，method的长度为1，认证的方式为0表示不认证 详见https://www.cnblogs.com/yinzhengjie/p/7357860.html
	licenseReq := clientLicenseReq{5, 1, [255]byte{0}}
	sendLicenseReq(conn, licenseReq)
	getLicenseResp(reader)
	//sock版本为5，3表示udp转发，rsv保留字段为0，1表示ipv4地址
	//ip只有在多主机的有意义，暂不关心，指定udp传输需要从哪一个端口开始发起
	connReq := clientConnReq{5, 3, 0, 1, [4]byte{0, 0, 0, 0}, [2]byte{4, 70}}
	sendClientConnReq(conn, connReq)
	proxyAddr, err := getConnResp(reader)
	if err != nil {
		fmt.Println(err)
	}
	//在传输UDP数据时，由于通过代理，所以需要按照一定的格式进行包装，在需要传送的数据之前添加一个报头，报头分析详见https://blog.csdn.net/petershina/article/details/9945615
	remsg, n, t := Send(proxyAddr, "www.google.com", "8.8.8.8", 53, false)
	fmt.Println(remsg, n, t)
}


