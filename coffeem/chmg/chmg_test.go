package chmg

import (
	"CoffeeM/coffeem/udp"
	"fmt"
	"log"
	"os"
	"testing"
)

type Addr struct {
	Net string
	Ip  string
}

func (addr Addr) Network() string {
	return addr.Net
}
func (addr Addr) String() string {
	return addr.Ip
}

func TestNormConditions(t *testing.T) {
	source := make(chan udp.Msg)
	defer close(source)
	destination := make(chan udp.Msg, 10)
	defer close(destination)
	lgr := log.New(os.Stderr, "chmg: ", log.LstdFlags)
	Start(destination, source, lgr)
	msg := udp.NewMsg()
	addr := Addr{
		Net: "udp4",
		Ip:  "192.0.2.1:25",
	}
	msg.Addr = addr
	fmt.Println("I'm Here")
	source <- *msg
	retMsg := <-destination
	fmt.Println(retMsg.Addr.String())
}
