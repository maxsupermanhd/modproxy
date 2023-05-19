package main

import (
	"log"
	"net/http"
	"os"
	"plugin"
	"strings"
)

func init() {
	dir, err := os.ReadDir(".")
	if err != nil {
		log.Println("Failed to read launch directory: ", err)
		return
	}
	for _, j := range dir {
		if !j.Type().IsRegular() {
			continue
		}
		if !strings.HasSuffix(j.Name(), ".so") {
			continue
		}
		p, err := plugin.Open(j.Name())
		if err != nil {
			log.Println("Error opening plugin: ", err)
			continue
		}
		ea, err := p.Lookup("ExportingActions")
		if err != nil {
			log.Println("Error getting exported actions: ", err)
			continue
		}
		eactions, ok := ea.(*map[string]func(w http.ResponseWriter, _ *http.Request, _ map[string]interface{}) error)
		if !ok {
			log.Println("Exported actions don't match type")
			continue
		}
		for k, v := range *eactions {
			log.Printf("Added action %q from plugin %q", k, j.Name())
			actionHandlers[k] = v
		}
	}
}
