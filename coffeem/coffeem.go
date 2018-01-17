package main

import (
	"CoffeeM/coffeem/udp"
	"log"
	"os"
)

func main() {
	cfg := udp.Config{
		IPListen: "127.0.0.1",
		Port:     10001,
	}
	msg := make(chan udp.Msg)
	lg := log.New(os.Stderr, "Pi:", log.LstdFlags)

	udp.Start(cfg)
}
