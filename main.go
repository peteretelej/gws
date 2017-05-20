package main

import (
	"flag"
	"log"
)

var (
	listen = flag.String("listen", "localhost:8080", "web server http listen address")
)

func main() {
	flag.Parse()

	svr := NewServer(*listen)
	log.Printf("Launching gws http server at %s", *listen)
	log.Fatal(svr.ListenAndServe())
}
