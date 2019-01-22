package main

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/evq/chromaticity/servers/api"
	"github.com/evq/chromaticity/servers/ssdp"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v1"
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
	app        = kingpin.New("chromaticity", "A Hue-like REST API for your lights")
	help       = app.Flag("help", "Show help.").Short('h').Dispatch(onHelp).Hidden().Bool()
	version    = app.Flag("version", "Show application version.").Short('v').Dispatch(onVersion).Bool()
	debug      = app.Flag("debug", "Enable debug mode.").Short('d').Bool()
	port       = app.Flag("port", "Api server port.").Short('p').Default("80").Int()
	configfile = app.Flag("config", "Config file.").Short('c').Default("").String()
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

	if *configfile == "" {
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			log.Error("Failed to determine homedir")
		}
		*configfile = homeDir + "/.chromaticity/data.json"
	}

	db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(errors.Wrap(err, "failed to connect database"))
	}
	defer db.Close()

	db.AutoMigrate(&chromaticity.ColorState{})

	ssdp.StartServer(strconv.Itoa(*port))
	api.StartServer(strconv.Itoa(*port), *configfile, db)
}
