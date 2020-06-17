package main

import (
	"flag"
	"log"
	"log/syslog"
	"net"
	"os"
	"strconv"
	"time"
)

func reflect(ip string, port, timeout, reset int) {
	log.Printf("Reflecting on %v:%v and logging inactivity > %vmsec.\n", ip, port, timeout)

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

	resetTimeout := time.Duration(reset) * time.Second
	lastReceived := time.Now()
	firstPacket := true
	buffer := make([]byte, 512)
	f0, f1 := 0, 1
	for {
		if time.Now().Sub(lastReceived) > resetTimeout {
			log.Printf("Exceeded silence deadline of %v seconds, resetting ...\n", reset)
			firstPacket = true
		}
		if firstPacket {
			conn.SetReadDeadline(time.Time{})
			log.Printf("Waiting for UDP stream ...\n")
		} else {
			nextTimeout := timeout + timeout*f0
			deadline := time.Now().Add(time.Duration(nextTimeout) * time.Millisecond)
			conn.SetReadDeadline(deadline)
		}
		n, remote, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
				log.Fatal(err)
			} else {
				log.Printf("Silence of %v msec.\n", time.Now().Sub(lastReceived).Milliseconds())
				f0, f1 = f1, f0+f1
			}
		} else {
			if firstPacket {
				log.Printf("Received first packet in stream.\n")
			}
			lastReceived = time.Now()
			_, err := conn.WriteToUDP(buffer[0:n], remote)
			if err != nil {
				log.Fatal(err)
			}
			firstPacket = false
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
	silenceTimeout := flag.Int("timeout", 2, "msec increments for timeout sequence")
	resetTimeout := flag.Int("reset", 120, "sec after which to reset for new stream")
	flag.Parse()

	reflect(*listenIP, *listenUDPPort, *silenceTimeout, *resetTimeout)
}
