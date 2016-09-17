package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/peteretelej/gws"
)

var (
	port = flag.Int("port", 8080, "port to listen on")
)

func main() {
	flag.Parse()
	listenAddr := fmt.Sprintf(":%d", *port)

	log.Printf("Launching gws http server at %s", listenAddr)
	log.Fatal(gws.Serve(listenAddr))
}
