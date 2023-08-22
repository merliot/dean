// This example listens on port :8080 for client connections.  Bytes
// received from the client are echo'ed back to the client.  Multiple
// clients can connect as the same time, each consuming a client socket,
// and being serviced by it's own go func.
//
// Example test using nc as client to copy file:
//
// $ nc 10.0.0.2 8080 <file >copy ; cmp file copy

//go:build pyportal || arduino_nano33 || nano_rp2040 || metro_m4_airlift || arduino_mkrwifi1010 || matrixportal_m4 || wioterminal

package main

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/merliot/dean/drivers/netlink"
	"github.com/merliot/dean/drivers/netlink/probe"
)

var (
	ssid string
	pass string
	port string = ":8080"
)

var buf [1024]byte

func echo(conn net.Conn) {
	defer conn.Close()
	_, err := io.CopyBuffer(conn, conn, buf[:])
	if err != nil && err != io.EOF {
		log.Fatal(err.Error())
	}
}

func main() {

	time.Sleep(time.Second)

	link, _ := probe.Probe()

	err := link.NetConnect(&netlink.ConnectParams{
		Ssid:       ssid,
		Passphrase: pass,
	})
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err.Error())
		}
		go echo(conn)
	}
}
