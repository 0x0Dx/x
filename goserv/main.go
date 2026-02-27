package main

import (
	"flag"
	"log"
	"net/http"
	"os"
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

	log.Printf("Serving %s on http://localhost:%s", *dir, *port)
	if err := http.ListenAndServe(":"+*port, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
