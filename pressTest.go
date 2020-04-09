package main

import (
	"fmt"
	"math/rand"
	"net"
	"sort"
	"time"
)

func pressTest(proxyServer string) {
	start_time := time.Now().UnixNano()
	wg.Add(assistNum)
	for i := 0; i < assistNum; i++ {
		go runsock5UdpForTest(assistQuery, proxyServer, start_time, queryTime, sockTimeOut, udpTimeOut, dnsServer, isProxy, isBlock, bufSize)
	}
	wg.Wait()
	end_time := time.Now().UnixNano()
	delayNum := len(delayList)
	sort.Sort(DelayList(delayList))
	var sumDelay int64 = 0
	for _, val := range delayList {
		sumDelay += val
	}

	fmt.Println("----------Request Count----------")
	fmt.Println("Total:       ", total)
	fmt.Println("Success:     ", success)
	fmt.Println("Failure:     ", failure)
	fmt.Println("SuccessRate: ", fmt.Sprintf("%.2f", ((success/total)*100.0)), "%")
	fmt.Println("------------QPS Count------------")
	fmt.Printf("Average QPS in %ds: %f \n", queryTime, float64(success)/float64(queryTime))
	fmt.Println("-----------Delay Count-----------")
	fmt.Printf("average delay:   %fms \n", float64(sumDelay)/float64(delayNum)/1e6)
	fmt.Printf("Minimum delay:   %fms \n", float64(delayList[0])/1e6)
	fmt.Printf("Maximum delay:   %fms \n", float64(delayList[delayNum-1])/1e6)
	fmt.Printf("Top 50%% delay:  %fms \n", float64(delayList[delayNum/2])/1e6)
	fmt.Printf("Top 75%% delay:  %fms \n", float64(delayList[delayNum*3/4])/1e6)
	fmt.Printf("Top 95%% delay:  %fms \n", float64(delayList[delayNum*95/100])/1e6)
	fmt.Printf("Top 99%% delay:  %fms \n", float64(delayList[delayNum*99/100])/1e6)
	fmt.Println("All query total usetime:", fmt.Sprintf("%.4f", float64(end_time-start_time)/1e9), "s")
}

func runsock5UdpForTest(num uint64, proxyServer string, t0, queryTime int64, sockTimeOut, udpTimeOut time.Duration, dnsServer string, isProxy, isBlock bool, bufSize uint32) {
	defer wg.Done()
	no := 0.0
	ok := 0.0
	var conn net.Conn
	var connUdp net.Conn
	var localDelay []int64
	conn, err := net.Dial("tcp", proxyServer)
	proxyAddr, err := sock5Auth(conn, proxyServer)
	//需要提前创建udp连接，只能开一个socket ，可以保证一次socket完成多次udp传送
	connUdp, err = net.DialTimeout("udp", proxyAddr, sockTimeOut)
	if err != nil {
		fmt.Println(err.Error())
	}
	buf := make([]byte, bufSize)
	for i := uint64(0); i < num; i++ {
		t1 := time.Now().UnixNano()
		if t1-t0 > queryTime*1e9 {
			break
		}
		_, _, t, err := sendUdp(&connUdp, &buf, dnsServer, isProxy, isBlock, udpTimeOut)
		if err != nil {
			fmt.Println(err)
			no += 1
			continue
		}
		ok += 1
		localDelay = append(localDelay, int64(t))
	}
	countMutex.Lock()
	failure += no
	success += ok
	total += no + ok
	countMutex.Unlock()
	delayMutex.Lock()
	delayList = append(delayList, localDelay...)
	delayMutex.Unlock()
	connUdp.Close()
	conn.Close()
}

func randDomainEvoke(len int) string {
	urlPre := "www."
	urlPost := ".com"
	urlBody := randSeq(len)
	return urlPre + urlBody + urlPost
}

func randSeq(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type DelayList []int64
func (DL DelayList) Len() int {
	return len(DL)
}

func (DL DelayList) Swap(i, j int) {
	DL[i], DL[j] = DL[j], DL[i]
}

func (DL DelayList) Less(i, j int) bool{ //sort from small to big
	return DL[i] < DL[j]
}
