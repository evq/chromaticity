package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/evq/chromaticity/servers/api"
	"github.com/evq/chromaticity/servers/ssdp"
	"gopkg.in/alecthomas/kingpin.v1"
	"os"
	"strconv"
)

func onHelp(context *kingpin.ParseContext) error {
	app.Usage(os.Stderr)
	os.Exit(0)
	return nil
}

func onVersion(context *kingpin.ParseContext) error {
	fmt.Println("0.0.1")
	os.Exit(0)
	return nil
}

var (
	app     = kingpin.New("chromaticity", "A Hue-like REST API for your lights")
	help    = app.Flag("help", "Show help.").Short('h').Dispatch(onHelp).Hidden().Bool()
	version = app.Flag("version", "Show application version.").Short('v').Dispatch(onVersion).Bool()
	debug   = app.Flag("debug", "Enable debug mode.").Short('d').Bool()
	port    = app.Flag("port", "Api server port.").Short('p').Default("80").Int()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	log.SetFormatter(&log.TextFormatter{
		ForceColors:      true,
		DisableColors:    false,
		DisableTimestamp: false,
		FullTimestamp:    true,
		TimestampFormat:  "",
		DisableSorting:   false,
	})
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	ssdp.StartServer(strconv.Itoa(*port))
	api.StartServer(strconv.Itoa(*port))
}
