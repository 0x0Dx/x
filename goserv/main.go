// goserv is a simple HTTP file server.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	dir  = flag.String("dir", ".", "directory to serve")
	port = flag.String("port", "8080", "port to listen on")
)

func main() {
	flag.Parse()

	if _, err := os.Stat(*dir); os.IsNotExist(err) {
		log.Fatalf("directory does not exist: %s", *dir)
	}

	handler := http.FileServer(http.Dir(*dir))

	srv := &http.Server{
		Addr:         ":" + *port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Serving %s on http://localhost:%s", *dir, *port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
