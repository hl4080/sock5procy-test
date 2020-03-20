package main


const (
	SIZEOFPROXYHEADER int = 10
	SIZEOFDNSHEADER int = 12
	SIZEOFDNSQUERY int = 4
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
	srv uint16 //保留两字节的0
	flag byte //是否数据报分段重组标志, 不分段为0
	atype byte //与clientConnReq相同
	addr uint32 //与clientConnReq相同？还是目标地址？
	port uint16 //	与clientConnReq相同？还是目标端口？
}