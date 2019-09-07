package main

import (
	"flag"

	"github.com/RyanJarv/gocash/gocash"
)

func main() {
	var ip = flag.String("ip", "127.0.0.1", "IP addr to bind and accept connections on")
	var port = flag.Int("port", 6379, "Port number to listen on")
	//var debug = flag.Bool("debug", true, "Run in foreground with debug output enabled")

	proc := gocash.NewProcess(
		gocash.NewServer(*ip, *port),
		gocash.NewCache(),
	)

	proc.Start()
}
