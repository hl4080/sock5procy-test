package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"
)

func sendLicenseReq(conn net.Conn, req clientLicenseReq) error {
	reqByte := []byte{req.ver, req.nmethods, req.methods[0]}
	conn.Write(reqByte)
	return nil
}

func getLicenseResp(r *bufio.Reader) error{
	log.Print("---------get license response from server---------")
	version, _ := r.ReadByte()
	log.Printf("protocol version is %d", version)
	if version != 5 {
		return errors.New("no sock5 protocol")
	}
	method, _:= r.ReadByte()
	log.Printf("method's length is %d", method)
	return nil
}

func sendClientConnReq(conn net.Conn, req clientConnReq) error {
	reqByte := []byte{req.ver, req.cmd, req.rsv, req.atyp}
	reqByte = append(reqByte, req.addr[0])
	reqByte = append(reqByte, req.addr[1])
	reqByte = append(reqByte, req.addr[2])
	reqByte = append(reqByte, req.addr[3])
	reqByte = append(reqByte, req.port[0])
	reqByte = append(reqByte, req.port[1])
	log.Printf("write dst ip : %s", getIpFromByte(req.addr, req.port))
	conn.Write(reqByte)
	return nil
}

func getConnResp(r *bufio.Reader) (string, error) {
	log.Print("---------get connection response from proxy server---------")
	version, _ := r.ReadByte()
	log.Printf("reconfirm protocol version is %d", version)
	if version != 5 {
		return "", errors.New("no sock5 protocol")
	}
	rep, _ := r.ReadByte()
	switch rep {
	case 0:
		log.Print("request connection succeed!")
	case 1:
		log.Print("server connection failed")
	case 2:
		log.Print("rules not allowed for connecting")
	case 3:
		log.Print("Internet not reachable")
	case 4:
		log.Print("host not reachable")
	case 5:
		log.Print("connection refused")
	case 6:
		log.Print("connection Timeout")
	case 7:
		log.Print("unsupport command")
	default:
		log.Print("undefined")
	}
	rsv, _ := r.ReadByte()
	log.Printf("rsv: %d", rsv)
	atype, _ := r.ReadByte()
	var addrLen byte
	addrLen = 0
	switch atype {
	case 1:
		log.Print("server get host as ipv4 address")
		addrLen = 4
	case 3:
		log.Print("server get host as domain name")
		addrLen, _ = r.ReadByte()
	case 4:
		log.Print("server get host as ipv6 address")
		addrLen = 6
	default:
		log.Print("not defined")
	}
	addr := make([]byte, addrLen)
	for i, _ := range(addr){
		addr[i], _ = r.ReadByte()
	}
	ip := fmt.Sprintf("%d.%d.%d.%d", addr[0], addr[1], addr[2], addr[3])
	log.Printf("get server's domain: %s", ip)
	//代理服务器根据自身资源返回指定的ip和端口用于客户端接受数据并转发
	var port uint16
	binary.Read(r, binary.BigEndian, &port)
	log.Printf("get server's port: %d", port)
	return fmt.Sprintf("%s:%d", ip, port), nil
}

func getIpFromByte(ipByte [4]byte, portByte [2]byte) string {
	ip := fmt.Sprintf("%d.%d.%d.%d", ipByte[0], ipByte[1], ipByte[2], ipByte[3])
	port := fmt.Sprintf("%d", int16(portByte[0])*256+int16(portByte[1]))
	return ip+" "+port
}

func ParseDomainName(domain string) []byte {
	var (
		buffer   bytes.Buffer
		segments []string = strings.Split(domain, ".")
	)
	for _, seg := range segments {
		binary.Write(&buffer, binary.BigEndian, byte(len(seg)))
		binary.Write(&buffer, binary.BigEndian, []byte(seg))
	}
	binary.Write(&buffer, binary.BigEndian, byte(0x00))

	return buffer.Bytes()
}

func ipStringToByte(stringIp string) uint32 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(stringIp).To4())
	ipint := ret.Int64()
	return uint32(ipint)
}

func (header *dnsHeader) SetFlag(QR uint16, OperationCode uint16, AuthoritativeAnswer uint16, Truncation uint16, RecursionDesired uint16, RecursionAvailable uint16, ResponseCode uint16) {
	header.Bits = QR<<15 + OperationCode<<11 + AuthoritativeAnswer<<10 + Truncation<<9 + RecursionDesired<<8 + RecursionAvailable<<7 + ResponseCode
}

func Send(conn net.Conn, domain string, dnsServer string, port uint16, isProxy bool) ([]byte, int, time.Duration, error) {
	pHeader := proxyHeader{
		srv:	0x0000,
		flag:	0,
		atype:	1,
		addr:	ipStringToByte(dnsServer),
		port:	port,
	}

	requestHeader := dnsHeader{
		Id:      0x0010,
		Qdcount: 1,
		Ancount: 0,
		Nscount: 0,
		Arcount: 0,
	}
	requestHeader.SetFlag(0, 0, 0, 0, 1, 0, 0)

	requestQuery := dnsQuery{
		QuestionType:  1,
		QuestionClass: 1,
	}

	var (
		err    error
		buffer bytes.Buffer
	)
	if isProxy {//在传输UDP数据时，由于通过代理，所以需要按照一定的格式进行包装，在需要传送的数据之前添加一个报头
		binary.Write(&buffer, binary.BigEndian, pHeader)
	}
	//defer conn.Close()
	binary.Write(&buffer, binary.BigEndian, requestHeader)
	binary.Write(&buffer, binary.BigEndian, ParseDomainName(domain))
	binary.Write(&buffer, binary.BigEndian, requestQuery)

	buf := make([]byte, 1024)
	t1 := time.Now()
	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		return make([]byte, 0), 0, 0, err
	}
	length, err := conn.Read(buf)
	if err != nil {
		return make([]byte, 0), 0, 0, err
	}
	t := time.Now().Sub(t1)
	return buf, length, t, nil
}

func sendUdp(connUdp net.Conn, testDomain, dnsServer string, isProxy bool) ([]byte, int, time.Duration, error) {
	var flag int
	for i, _ := range dnsServer{
		if dnsServer[i] == ':'{
			flag = i
			break
		}
	}
	dnsIp := dnsServer[0:flag]
	dnsPortInt, err := strconv.Atoi(dnsServer[flag+1:])
	if err != nil {
		log.Println(err)
	}
	dnsPort := uint16(dnsPortInt)
	remsg, n, t, err := Send(connUdp, testDomain, dnsIp, dnsPort, isProxy)
	return remsg, n, t, err
}

func sock5Auth(conn net.Conn, proxyServer string) (string, error) {
	//通过tcp连接到sock5的udp代理与代理进行验证
	/*conn, err := net.Dial("tcp", proxyServer)
	if err != nil {
		return "", err
	}*/
	var err error
	fmt.Printf("获取%s的tcp连接成功...\n", proxyServer)
	reader := bufio.NewReader(conn)
	//协商版本sock5，method的长度为1，认证的方式为0表示不认证
	licenseReq := clientLicenseReq{5, 1, [255]byte{0}}
	err = sendLicenseReq(conn, licenseReq)
	if err != nil {
		return "", err
	}
	err = getLicenseResp(reader)
	if err != nil {
		return "", err
	}
	//sock版本为5，3表示udp转发，rsv保留字段为0，1表示ipv4地址
	//ip只有在多主机的有意义，暂不关心
	connReq := clientConnReq{5, 3, 0, 1, [4]byte{0, 0, 0, 0}, [2]byte{0, 0}}
	err = sendClientConnReq(conn, connReq)
	if err != nil {
		return "", err
	}
	proxyAddr, err := getConnResp(reader)
	if err != nil {
		return "", err
	}
	return proxyAddr, nil
}