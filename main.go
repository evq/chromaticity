package main

import (
	"github.com/evq/chromaticity/servers/api"
	"github.com/evq/chromaticity/servers/ssdp"
)

const API_PORT = "80"

func main() {
  ssdp.StartServer(API_PORT)
	api.StartServer(API_PORT)
}
