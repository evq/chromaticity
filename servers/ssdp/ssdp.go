package ssdp

import (
	"bytes"
	"net"
	"strings"
	"text/template"
)

type Service struct {
	IP   string
	Port string
}

const (
	UPNP_ADDR     = "239.255.255.250:1900"
	UPNP_RESPONSE = `HTTP/1.1 
CACHE-CONTROL: max-age=100
EXT:
LOCATION: http://{{.IP}}:{{.Port}}/description.xml
SERVER: FreeRTOS/6.0.5, UPnP/1.0, IpBridge/0.1
ST: upnp:rootdevice
USN: uuid:2f402f80-da50-11e1-9b23-0017880967e7
`
)

var UpnpAddr *net.UDPAddr
var T *template.Template

func init() {
	var err error
	UpnpAddr, err = net.ResolveUDPAddr("udp4", UPNP_ADDR)
	if err != nil {
		panic(err)
	}
	T = template.New("ssdp")
	T, _ = T.Parse(UPNP_RESPONSE)
}

func StartServer(port string) {
	go ListenAndRespond(port)
}

func ListenAndRespond(port string) {
	conn, err := net.ListenMulticastUDP("udp", nil, UpnpAddr)
	if err != nil {
		panic(err)
	}
	for {
		buf := make([]byte, 1024)
		n, sender, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}
		if strings.Contains(string(buf[:n]), "M-SEARCH") {
			Resp := bytes.Buffer{}
			ifaces, _ := net.Interfaces()
			for _, iface := range ifaces {
				addrs, _ := iface.Addrs()
				for _, addr := range addrs {
					switch data := addr.(type) {
					case *net.IPNet:
						if data.Contains(sender.IP) {
							T.Execute(&Resp, Service{data.IP.String(), port})
							conn.WriteToUDP(Resp.Bytes(), sender)
							break
						}
					}
				}
			}
		}
	}
}
