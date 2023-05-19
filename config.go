package main

import (
	"log"
	"os"

	"github.com/maxsupermanhd/lac"
)

var (
	cf *lac.Conf
)

func loadConfig() {
	configPath := "config.json"
	if os.Getenv("CONFIG_PATH") != "" {
		configPath = os.Getenv("CONFIG_PATH")
	}
	var err error
	cf, err = lac.FromFileJSON(configPath)
	if err != nil {
		log.Printf("Failed to load configuration (%v), defaults will be used!", err)
		cf = lac.NewConf()
	}
}
