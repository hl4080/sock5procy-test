package main

import (
	"sync"
	"time"
)

var (
	total   	float64 		= 0.0
	success 	float64 		= 0.0
	failure 	float64 		= 0.0
	assistNum	int 			= 100  //并发请求的协程数
	assistQuery	uint64			= 1844674407370955161		//保证每个协程内的数据能一直收发
	queryTime	int64			= 60	//持续请求的时间
	isProxy 	bool 			= false 	//是否通过代理进行dns 查询
	isBlock		bool			= false	//udp大块传输
	proxyServer string 			= "10.20.47.145:1090" 	//sock5代理服务器的ip地址
	dnsServer 	string 			= "203.119.159.121:53" 	//需要请求的dns server的ip地址
	domain 		string 			= "www.byted.online" 	//想知道ip的域名地址
	bufSize		uint32			= 1024					//dns收发包的缓存大小
	sockTimeOut	time.Duration	= 80*time.Second 		//socket超时设置
	udpTimeOut	time.Duration	= 800*time.Millisecond 	//udp超时设置
	delayList	[]int64
	wg			sync.WaitGroup
	countMutex 	sync.Mutex
	delayMutex	sync.Mutex
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

type dnsAdditional struct {
	Name 	uint8
	Type 	uint16
	Payload uint16
	Extend 	uint8
	Version uint8
	Z		uint16
	DataLen uint16
}

type proxyHeader struct {
	srv uint16 	//保留两字节的0
	flag byte 	//是否数据报分段重组标志, 不分段为0
	atype byte 	//与clientConnReq相同
	addr uint32 //目标dns服务器的地址
	port uint16 //目标dns服务器的端口
}