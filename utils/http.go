package utils

import (
	"net"
	"net/http"
	"strings"
)

func GetHostPort(req *http.Request) (string, string) {
	var IP string
	var Port string
	if !strings.Contains(req.Host,":") {
		IP = req.Host
		Port = "80"
	} else {
		split := strings.Split(req.Host,":")
		IP = split[0]
		Port = split[1]
	}
	ipaddr, _ := net.ResolveIPAddr("ip4", IP)
	return ipaddr.String(), Port
}

// Returns a "dummy" gateway ( 
func DummyGateway(host string) string {
	ipseg := strings.Split(host, ".")
	ipseg[3] = "1"
	return strings.Join(ipseg, ".")
}
