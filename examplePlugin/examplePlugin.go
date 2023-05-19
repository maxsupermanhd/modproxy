package main

import (
	"log"
	"net/http"
)

var (
	ExportingActions = map[string]func(w http.ResponseWriter, _ *http.Request, _ map[string]interface{}) error{
		"testAction": func(_ http.ResponseWriter, r *http.Request, _ map[string]interface{}) error {
			log.Println("Exported action performed on ", r.RequestURI)
			return nil
		},
	}
)
