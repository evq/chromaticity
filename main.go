package main

import (
	"github.com/evq/chromaticity/apiserver"
	"log"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)

	apiserver.StartApiServer()
}
