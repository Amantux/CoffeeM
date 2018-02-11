package main

import (
	"CoffeeM/coffeem/tcp"
	"CoffeeM/coffeem/udp"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
)

func msgPrint(msg <-chan udp.Msg, wg *sync.WaitGroup) {
	for m := range msg {
		fmt.Println(m.Pld)
	}
	wg.Done()
}
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}
func AddChannel(TargetChan chan string, str string) {
	fmt.Println("Pree add")
	fmt.Println(str)
	TargetChan <- str
	fmt.Println("Added")
}

func main() {
	cfg := tcp.Config{
		IPListen: "127.0.0.1",
		Port:     10001,
	}
	msg := make(chan tcp.Msg)
	defer close(msg)
	lg := log.New(os.Stderr, "Pi:", log.LstdFlags)
	term := make(chan bool)
	go termRasPI(term)
	status := make(chan string)
	defer close(status)
	var wg sync.WaitGroup
	udp.Start(cfg, msg, lg, &wg, status, term)
	<-term
	wg.Wait()
}

func termRasPI(term chan bool) {
	osSig := make(chan os.Signal)
	defer close(osSig)
	defer close(t)
	// Currently listens for only SIGTERM as it's available on every
	// flavor of Linux.
	signal.Notify(osSig, os.Interrupt)
	// blocks until it receives at least one signal
	<-osSig
}
