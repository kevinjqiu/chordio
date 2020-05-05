package server

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
)

func getAvailableInterfaces() []net.Interface {
	var ai = make([]net.Interface, 0, 0)
	interfaces, _:= net.Interfaces()
	for _, inf := range interfaces {
		isBroadcast := (inf.Flags & net.FlagBroadcast) == net.FlagBroadcast
		isUp := (inf.Flags & net.FlagUp) == net.FlagUp
		if isBroadcast && isUp {
			ai = append(ai, inf)
		}
	}
	return ai
}

func isIPv4(addr net.Addr) bool {
	parts := strings.Split(addr.String(), ".")
	i, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	return i >= 0 && i < 256
}

func canBindIP(addr string) bool {
	ips, err := net.InterfaceAddrs()
	if err != nil {
		logrus.Error("cannot enumerate interface IPs: ", err)
		return false
	}

	for _, ip := range ips {
		if strings.Split(ip.String(), "/")[0] == addr {
			return true
		}
	}
	return false
}

func getFirstAvailableBindIP() (string, error) {
	ifs := getAvailableInterfaces()

	for _, inf := range ifs {
		addrs, err := inf.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			if isIPv4(addr) {
				return strings.Split(addr.String(), "/")[0], nil
			}
		}
	}

	return "", fmt.Errorf("unable to get available IP")
}
