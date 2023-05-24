package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var (
	mon = monitor{
		e:        make(chan logRequest, 10),
		newConns: make(chan chan logRequest, 5),
		remConns: make(chan chan logRequest, 5),
		stop:     make(chan struct{}),
	}
)

type logRequest struct {
	Err            string
	ResponseCode   int
	UserAgent      string
	Path           string
	SourceIP       string
	When           string
	ProcessingTime string
}

type monitor struct {
	e        chan logRequest
	newConns chan chan logRequest
	remConns chan chan logRequest
	stop     chan struct{}
}

func (m monitor) Run() {
	conns := map[chan logRequest]bool{}
	for {
		select {
		case l := <-m.e:
			for c := range conns {
				c <- l
			}
		case n := <-m.newConns:
			conns[n] = true
		case r := <-m.remConns:
			delete(conns, r)
			close(r)
		case <-m.stop:
			for c := range conns {
				close(c)
			}
			return
		}
	}
}

func init() {
	go mon.Run()
}

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	ind, err := os.ReadFile("index.html")
	if err != nil {
		w.WriteHeader(500)
	}
	w.WriteHeader(200)
	w.Write(ind)
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/event-stream")
	c := make(chan logRequest, 8)
	mon.newConns <- c
	go func() {
		<-r.Context().Done()
		mon.remConns <- c
	}()
	w.WriteHeader(200)
	fmt.Fprint(w, "data: connected\n\n")
	f.Flush()
	for m := range c {
		b, err := json.Marshal(m)
		if err != nil {
			fmt.Fprintf(w, "data: %v\n\n", err)
			continue
		}
		fmt.Fprintf(w, "data: %s\n\n", b)
		f.Flush()
	}
}
