package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	log.Println("ModProxy starting up...")

	log.Println("Reading config...")
	loadConfig()

	log.Println("Preparing HTTP multiplexer...")
	proxyMux := &Mux{}
	controlMux := http.NewServeMux()
	controlMux.HandleFunc("/", indexHandler)
	controlMux.HandleFunc("/sse", sseHandler)

	log.Println("Starting HTTP servers...")
	ctx, ShutdownFunc := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	startHTTP(&wg, ctx, cf.GetDString("0.0.0.0:8030", "proxy", "listen"), proxyMux)
	startHTTP(&wg, ctx, cf.GetDString("0.0.0.0:8031", "listenControl"), controlMux)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	ShutdownFunc()
	mon.stop <- struct{}{}
	log.Println("Shutdown issued")
	wg.Wait()
	log.Println("Bye!")
}

func startHTTP(wg *sync.WaitGroup, ctx context.Context, addr string, mux http.Handler) {
	wg.Add(1)
	go func() {
		srv := http.Server{
			Addr:    addr,
			Handler: mux,
		}
		wg.Add(1)
		go func() {
			log.Printf("Starting listening for HTTP requests on %q", addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Web server returned an error: %s\n", err)
			}
			wg.Done()
		}()
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Server Shutdown Failed:%+v", err)
		} else {
			log.Println("Web server stopped")
		}
		wg.Done()
	}()
}
