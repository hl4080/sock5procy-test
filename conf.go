package main

import (
	"sync"
	"time"
)

var (
	total   	float64 		= 0.0
	success 	float64 		= 0.0
	failure 	float64 		= 0.0
	assistNum	int 			= 2000  //并发请求的协程数
	assistTime	int				= 2		//每个协程请求的udp数
	isProxy 	bool 			= true 					//是否通过代理进行dns 查询
	proxyServer string 			= "10.20.47.145:1080" 	//sock5代理服务器的ip地址
	dnsServer 	string 			= "116.211.173.141:53" 	//需要请求的dns server的ip地址
	domain 		string 			= "www.byted.online" 		//想知道ip的域名地址
	udpTimeOut	time.Duration	= 2000 	//udp超时设置800ms
	wg			sync.WaitGroup
	mutex 		sync.Mutex
)

type clientLicenseReq struct {
	ver byte
	nmethods byte
	methods [255]byte
}
type serverLicenseResp struct {
	ver byte
	methods byte
}
type clientConnReq struct {
	ver byte
	cmd byte
	rsv byte
	atyp byte
	addr [4]byte
	port [2]byte
}
type serverConnResp struct {
	ver byte
	res byte
	rsv byte
	atype byte
	addr [4]byte
	port [2]byte
}

type dnsHeader struct {
	Id                                 uint16
	Bits                               uint16
	Qdcount, Ancount, Nscount, Arcount uint16
}

type dnsQuery struct {
	QuestionType  uint16
	QuestionClass uint16
}

type proxyHeader struct {
	srv uint16 	//保留两字节的0
	flag byte 	//是否数据报分段重组标志, 不分段为0
	atype byte 	//与clientConnReq相同
	addr uint32 //目标dns服务器的地址
	port uint16 //目标dns服务器的端口
}