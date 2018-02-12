package tcp

import (
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"sync"
	"time"
)

type Config struct {
	IPListen string
	Port     int
}

const PktSz = 1024 //units: byte

type Msg struct {
	Addr net.Addr
	Pld  []byte
}

func NewMsg() (m *Msg) {
	m = new(Msg)
	m.Pld = make([]byte, PktSz)
	return
}

var lg *log.Logger

func Start(
	cfg Config,
	msg chan<- Msg,
	lgr *log.Logger,
	wg *sync.WaitGroup,
	status chan<- string,
	term <-chan bool,
) (
	err error,
) {
	lg = lgr
	addr := net.TCPAddr{IP: net.ParseIP(cfg.IPListen),
		Port: cfg.Port,
	}
	var l *net.TCPListener
	if l, err = net.ListenTCP("tcp", &addr); err != nil {
		return
	}
	lg.Print("hi there!")
	trig := make(chan bool)
	defer close(trig)
	wg.Add(1)
	go connManager(l, msg, trig, wg, status, term)
	// block till connManager ready to accept messages
	<-trig
	return
}

const deadLineInterval = 10 * time.Minute

func connManager(
	l *net.TCPListener,
	msg chan<- Msg,
	trig chan<- bool,
	wg *sync.WaitGroup,
	status chan<- string,
	term <-chan bool,
) {
	defer wg.Done()
	wg.Add(1)
	go termListener(l, wg, term)
	// signal connection manager started
	trig <- true
	const retryMax = 6
	var retryAttempts int
	for {
		retryAttempts = 0
	connRetry:
		retryAttempts += 1
		conn, err := l.AcceptTCP()
		if err != nil {
			if eofDetected, _ := regexp.MatchString(".*closed network connection.*", err.Error()); eofDetected {
				return
			}
			if retryAttempts < retryMax {
				lg.Printf("Retry Attempt: %d of %d\n", retryAttempts, retryMax)
				time.Sleep(time.Second * 2) //Gives it time to get its shit together
				goto connRetry
			}
			status <- "Failure"
			lg.Printf("Error failed after %d retries. Connection: %v\n", retryAttempts, conn)
			status <- "Failure"
			return
		}
		fmt.Fprintf(os.Stderr, "after accept \n")

		if err := conn.SetReadDeadline(time.Now().Add(deadLineInterval)); err != nil {
			lg.Printf("Error Read Deadline: %s \n", err.Error())
			status <- "Failure"
			return
		}
		if err := conn.SetReadBuffer(PktSz); err != nil {
			lg.Printf("Error Read Buffer: %s \n", err.Error())
			status <- "Failure"
			return
		}
		if err := conn.SetWriteBuffer(PktSz); err != nil {
			lg.Printf("Error Write Buffer: %s \n", err.Error())
			status <- "Failure"
			return
		}
		wg.Add(1)
		go readMessage(conn, msg, wg, term)
	}
}
func readMessage(
	conn *net.TCPConn,
	msg chan<- Msg,
	wg *sync.WaitGroup,
	term <-chan bool,
) {
	defer wg.Done()
	wg.Add(1)
	go closeConn(conn, wg, term)
	for {
		m := NewMsg()
		if len, err := conn.Read(m.Pld); err != nil {
			if closed, _ := regexp.MatchString(".*closed network connection.*|.*EOF.*", err.Error()); closed {
				return
			}
			lg.Printf("Error Reading: %s size of %d \n", err.Error(), len)
		}
		if err := conn.SetReadDeadline(time.Now().Add(deadLineInterval)); err != nil {
			lg.Printf("Error Read Deadline: %s \n", err.Error())
			return
		}
		fmt.Fprintf(os.Stderr, "In read message.\n")
		msg <- *m
	}
}
func termListener(
	l *net.TCPListener,
	wg *sync.WaitGroup,
	term <-chan bool,
) {
	<-term
	l.Close()
	wg.Done()
}
func closeConn(
	conn *net.TCPConn,
	wg *sync.WaitGroup,
	term <-chan bool,
) {
	<-term
	conn.Close()
	wg.Done()
}
