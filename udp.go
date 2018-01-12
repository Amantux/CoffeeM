package udp

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

type Config struct {
	IPListen string
	Port     int
}

type Msg struct {
	Addr net.Addr
	Pld  []byte
}

func NewMsg() (m *Msg) {
	const pktSz = 1024
	m = new(Msg)
	m.Pld = make([]byte, pktSz)
	return
}

var lg *log.Logger

func Start(
	cfg Config,
	msg chan<- Msg,
	lgr *log.Logger,
	wg *sync.WaitGroup,
	term <-chan bool,
) (
	err error,
) {
	lg = lgr
	addr := net.UDPAddr{IP: net.ParseIP(cfg.IPListen),
		Port: cfg.Port,
	}
	var c *net.UDPConn
	if c, err = net.ListenUDP("udp4", &addr); err != nil {
		return
	}
	trig := make(chan bool)
	defer close(trig)
	wg.Add(1)
	go listen(c, msg, trig, wg, term)
	<-trig
	return
}

func listen(
	c *net.UDPConn,
	msg chan<- Msg,
	trig chan<- bool,
	wg *sync.WaitGroup,
	term <-chan bool,
) {
	defer wg.Done()
	defer c.Close()
	wg.Add(1)
	// UDP read blocks :: start in goroutine
	go readBlock(c, msg, trig, wg, term)
	lg.Println("Monitoring connection")
	<-term
	fmt.Fprintln(os.Stderr, "terminated listener")
	// Close connection which causes read to fail terminating
	// the read goroutine
}

func readBlock(
	c *net.UDPConn,
	msg chan<- Msg,
	trig chan<- bool,
	wg *sync.WaitGroup,
	term <-chan bool,
) {
	defer wg.Done()
	var err error
	trig <- true
	for {
		m := NewMsg()
		var pldLen int
		pldLen, m.Addr, err = c.ReadFrom(m.Pld)
		select {
		case <-term:
			fmt.Fprintln(os.Stderr, "closed readBlock")
			return
		default:
			// haven't terminated server
			if err != nil {
				// encountered read error
				lg.Println(err.Error())
				return
			}
			// no errors && server active :: forward message
			lg.Print("Received message")
			m.Pld = m.Pld[:pldLen]
			msg <- *m
		}
	}
}
