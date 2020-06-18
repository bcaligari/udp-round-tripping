package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func ping(target string, port, interval, duration int) {
	log.Printf("Pinging %v:%v with %v microsecond intervals for %v seconds.\n",
		target, port, interval, duration)

	udpEndPoint := target + ":" + strconv.Itoa(port)
	nanoInterval := time.Duration(interval) * time.Microsecond
	startTime := time.Now()
	quitAfter := time.Duration(duration) * time.Second
	udpAddr, err := net.ResolveUDPAddr("udp4", udpEndPoint)
	var totalSent, totalReceived, errorsWrite, errorsRead int64
	if err != nil {
		log.Fatal(udpEndPoint + " is not a valid UDP endpoint.")
	}
	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	buffer := make([]byte, 2048)

	for {
		if quitAfter > 0 && time.Now().Sub(startTime) > quitAfter {
			log.Printf("Reached the end of my %v second run.\n", duration)
			log.Printf("Total sent: %d, total received: %d\n", totalSent, totalReceived)
			log.Printf("Write errors: %d, read errors: %d\n", errorsWrite, errorsRead)
			return
		}
		udpTimeStamp := time.Now().UnixNano()
		fmt.Printf(">> %d\n", udpTimeStamp)
		timestamp := make([]byte, 8)
		binary.LittleEndian.PutUint64(timestamp, uint64(udpTimeStamp))
		_, err := conn.Write(timestamp)
		totalSent++
		if err != nil {
			errorsWrite++
		}
		deadline := time.Now().Add(time.Duration(nanoInterval))
		conn.SetReadDeadline(deadline)
		for {
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				if e, ok := err.(net.Error); ok && e.Timeout() {
					break
				} else {
					errorsRead++
				}
				log.Println(err)
			} else {
				totalReceived++
				fmt.Printf("<< %d\n", int64(binary.LittleEndian.Uint64(buffer[0:n])))
			}
		}
	}
}

func main() {
	targetServer := flag.String("target", "127.0.0.1", "Target IP to ping with UDP traffic")
	targetUDPPort := flag.Int("port", 36000, "UDP port to send traffic on")
	sendInterval := flag.Int("interval", 500, "microsecond interval between UDP sends")
	quitAfter := flag.Int("quit", 300, "seconds after which to quit")
	logToSyslog := flag.Bool("syslog", false, "Log to syslog")
	flag.Parse()

	if *logToSyslog {
		logwriter, err := syslog.New(syslog.LOG_INFO, filepath.Base(os.Args[0]))
		if err == nil {
			log.SetOutput(logwriter)
		}
	}

	ping(*targetServer, *targetUDPPort, *sendInterval, *quitAfter)
}
