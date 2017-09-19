// +build linux

package utils

import (
	"log"
	"net"
	"runtime"

	"golang.org/x/net/ipv4"
)

func ListenUdpMulticast(addr *net.UDPAddr) (*net.UDPConn, error) {
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if runtime.GOOS == "linux" {

		pc := ipv4.NewPacketConn(conn)

		ifaces, err := net.Interfaces()
		if err != nil {
			log.Println("can't get interface list")
			return nil, err
		}

		for _, iface := range ifaces {
			if err = pc.JoinGroup(&iface, &net.UDPAddr{IP: addr.IP}); err != nil {
				continue
			}

			// test
			if loop, err := pc.MulticastLoopback(); err == nil {
				//fmt.Printf("MulticastLoopback status:%v\n", loop)
				if !loop {
					pc.SetMulticastLoopback(true)
				}
			}
		}
	}
	return conn, err
}
