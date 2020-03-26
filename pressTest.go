package main

import (
	"fmt"
	"net"
	"time"
)

func pressTest(isProxy bool, proxyServer, dnsServer, domain string) {
	start_time := time.Now().UnixNano()
	wg.Add(assistNum)
	for i := 0; i < assistNum; i++ {
		go runsock5UdpForTest(assistTime, isProxy, proxyServer, dnsServer, domain)
	}
	wg.Wait()
	end_time := time.Now().UnixNano()
	fmt.Println("Total:", total)
	fmt.Println("Success:", success)
	fmt.Println("Failure:", failure)
	fmt.Println("SuccessRate:", fmt.Sprintf("%.2f", ((success/total)*100.0)), "%")
	fmt.Println("UseTime:", fmt.Sprintf("%.4f", float64(end_time-start_time)/1e9), "s")
}

func runsock5UdpForTest(num int, isProxy bool, proxyServer, dnsServer, domain string) {
	defer wg.Done()
	no := 0.0
	ok := 0.0
	var conn net.Conn
	var connUdp net.Conn
	conn, err := net.Dial("tcp", proxyServer)
	//conn.SetReadDeadline(time.Now().Add(udpTimeOut * time.Millisecond))
	proxyAddr, err := sock5Auth(conn, proxyServer)
	//需要提前创建udp连接，只能开一个socket ，可以保证一次socket完成多次udp传送
	connUdp, err = net.Dial("udp", proxyAddr)
	connUdp.SetReadDeadline(time.Now().Add(udpTimeOut * time.Millisecond)) //强行结束udp数据传输时的长时间阻塞情况
	if err != nil {
		fmt.Println(err.Error())
	}
	t0 := time.Now().UnixNano()
	for i := 0; i < num; i++ {
		_, _, _, err = sendUdp(connUdp, domain, dnsServer, isProxy)
		if err != nil {
			fmt.Println(err)
			no += 1
			continue
		}
		if time.Now().UnixNano() - t0 <= 1e9 { //同一秒内完成的数量
			ok += 1
		}
	}
	mutex.Lock()
	failure += no
	success += ok
	total += float64(num)
	mutex.Unlock()
	connUdp.Close()
	conn.Close()
}
