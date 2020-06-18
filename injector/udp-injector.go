package main

import (
	"encoding/binary"
	"flag"
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
	nanoInterval := time.Duration(interval * 1000)
	startTime := time.Now()
	quitAfter := time.Duration(duration) * time.Second
	udpAddr, err := net.ResolveUDPAddr("udp4", udpEndPoint)
	if err != nil {
		log.Fatal(udpEndPoint + " is not a valid UDP endpoint.")
	}
	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for {
		if quitAfter > 0 && time.Now().Sub(startTime) > quitAfter {
			log.Printf("Reached the end of my %v second run.\n", duration)
			return
		}
		timestamp := make([]byte, 8)
		binary.LittleEndian.PutUint64(timestamp, uint64(time.Now().UnixNano()))
		conn.Write(timestamp)
		time.Sleep(nanoInterval)
	}

}

func main() {
	targetServer := flag.String("target", "127.0.0.1", "Target IP to ping with UDP traffic")
	targetUDPPort := flag.Int("port", 36000, "UDP port to send traffic on")
	sendInterval := flag.Int("timeout", 500, "microsecond interval between UDP sends")
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
