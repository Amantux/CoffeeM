package main

import (
	"CoffeeM/coffeem/tcp"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
)

func main() {
	cfg := tcp.Config{
		IPListen: "0.0.0.0",
		Port:     10001,
	}
	lg := log.New(os.Stderr, "Pi:", log.LstdFlags)
	term := make(chan bool)
	go termRasPI(term)
	status := make(chan string)
	defer close(status)
	var wg sync.WaitGroup
	msgOut, err := tcp.Start(cfg, lg, &wg, status, term)
	if err != nil {
		lg.Fatalf("Fatal: Could not start server: '%s'.", err.Error())
		os.Exit(1)
	}
	queueOut := startQueue(msgOut, lg)
	relayOut := startRelay(queueOut, lg)
	//go msgPrint(msgOut, lg, &wg)
	// coordinate graceful terination
	select {
	case <-term:
	case <-status:
		// suicide call
		termMyself()
		<-term
	}
	<-relayOut
	wg.Wait()
}

func startQueue(msgIn <-chan tcp.Msg, lg *log.Logger) <-chan tcp.Msg {
	msgChan := make(chan tcp.Msg, 20)
	go queue(msgIn, msgChan, lg)
	return msgChan
}

func queue(msgIn <-chan tcp.Msg, msgChan chan tcp.Msg, lg *log.Logger) {
	defer close(msgChan)
	for {
		select {
		case msg, open := <-msgIn:
			if !open {
				return
			}
			select {
			case msgChan <- msg:
			default:
				lg.Printf("Error: Exceeded queue size: processing message:'%v'", msg)
			}
		}
	}
}

func startRelay(msgIn <-chan tcp.Msg, lg *log.Logger) <-chan bool {
	killChan := make(chan bool)
	go relay(msgIn, killChan, lg)
	return killChan
}

func relay(msg <-chan tcp.Msg, killChan chan<- bool, lg *log.Logger) {
	defer close(killChan)
	var arduino *tcp.Msg
	var Nexus *tcp.Msg
	for m := range msg {
		payload := string(m.Pld)
		lg.Printf("Info: Received Message:'%s'.", payload)
		if strings.HasPrefix(payload, "Arduino") {
			arduino = &m
		} else if strings.HasPrefix(payload, "Nexus") && Nexus != nil {
			Nexus = &m
		}
		if Nexus != nil && arduino != nil {
			err := arduino.Reply(Nexus.Pld)
			if err != nil {
				lg.Printf("Error: line 49: '%s'.", err.Error())
			}
		}
	}
}

func msgPrint(msg <-chan tcp.Msg, lg *log.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	for m := range msg {
		lg.Printf("Info: Received message: '%s'.", string(m.Pld))
	}
}

func termMyself() {
	pid := os.Getpid()
	mySelf, _ := os.FindProcess(pid)
	mySelf.Signal(os.Interrupt)
}

func termRasPI(term chan bool) {

	osSig := make(chan os.Signal)
	defer close(osSig)
	defer close(term)
	// Currently listens for only SIGTERM as it's available on every
	// flavor of Linux.
	signal.Notify(osSig, os.Interrupt)
	// blocks until it receives at least one signal
	<-osSig
	signal.Stop(osSig)
}
