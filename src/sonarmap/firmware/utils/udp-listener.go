// +build !linux

package utils

import (
	"log"
	"net"
	"runtime"

	"golang.org/x/net/ipv4"
)

func ListenUdpMulticast(addr *net.UDPAddr) (conn *net.UDPConn, err error) {
	conn, err = net.ListenMulticastUDP("udp", nil, addr)
	return conn, err
}
