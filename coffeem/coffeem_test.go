package main

import (
	"CoffeeM/coffeem/tcp"
	"log"
	"os"
	"sync"
	"testing"
)

func TestMessageQueue(t *testing.T) {
	lg := log.New(os.Stderr, "Pi:", log.LstdFlags)
	Input := make(chan tcp.Msg)
	Output := startQueue(Input, lg)
	var wg sync.WaitGroup
	wg.Add(1)
	go msgPrint(Output, lg, &wg)
	Message := tcp.NewMsg()
	MessageString := "This is a test"
	Message.Pld = []byte(MessageString)
	Input <- *Message
	close(Input)
	wg.Wait()
}
func TestMessageOverflow(t *testing.T) {
	lg := log.New(os.Stderr, "Pi:", log.LstdFlags)
	Input := make(chan tcp.Msg)
	Output := startQueue(Input, lg)
	//var wg sync.WaitGroup
	for i := 0; i < 21; i++ {
		Message := tcp.NewMsg()
		MessageString := "This is a test"
		Message.Pld = []byte(MessageString)
		Input <- *Message
	}
	close(Input)
	for {
		_, open := <-Output
		if !open {
			break
		}
	}
}
func TestRelay(t *testing.T) {
	lg := log.New(os.Stderr, "Relay:", log.LstdFlags)
	Input := make(chan tcp.Msg)
}
