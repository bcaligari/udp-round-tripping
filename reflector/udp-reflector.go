package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"
	"strconv"
	"time"
)

func reflect(ip string, port, timeout int) {
	log.Printf("Reflecting on %v:%v and logging inactivity > %vmsec\n", ip, port, timeout)
	udpEndPoint := ip + ":" + strconv.Itoa(port)

	udpAddr, err := net.ResolveUDPAddr("udp4", udpEndPoint)
	if err != nil {
		log.Fatal(udpEndPoint + " is not a valid UDP endpoint.")
	}

	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buffer := make([]byte, 512)
	for {
		deadline := time.Now().Add(time.Duration(timeout) * time.Millisecond)
		conn.SetReadDeadline(deadline)
		n, remote, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
				log.Fatal(err)
			} else {
				log.Printf("Timeout ...")
			}
		} else {
			_, err = conn.WriteToUDP(buffer[0:n], remote)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func main() {
	logwriter, err := syslog.New(syslog.LOG_INFO, os.Args[0])
	if err == nil {
		log.SetOutput(logwriter)
	}
	listenIP := flag.String("ip", "0.0.0.0", "Target IP to ping with UDP traffic")
	listenUDPPort := flag.Int("port", 36000, "UDP port to send traffic on")
	silenceTimeout := flag.Int("timeout", 3000, "msec inactivity to log timeout")
	flag.Parse()

	reflect(*listenIP, *listenUDPPort, *silenceTimeout)
	fmt.Println("...")
}
