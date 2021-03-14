package main

import (
	"flag"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()
	server := NewServer()
	server.Run(*addr)
}
