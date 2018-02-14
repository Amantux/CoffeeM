package tcp

import (
	"fmt"
	"log"
	"net"
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
	Addr      net.Addr
	Pld       []byte
	reply     chan Msg
	replyTOut time.Duration
}

func NewMsg() (m *Msg) {
	m = new(Msg)
	m.Pld = make([]byte, PktSz)
	return
}
func (m Msg) Reply(pld []byte) (err error) {
	if m.reply == nil {
		err = fmt.Errorf("No connection to issue to reply.")
		return
	}
	tout := m.replyTOut
	if m.replyTOut == 0 {
		tout = 1 * time.Second
	}
	select {
	case m.reply <- m:
	case <-time.After(tout):
		err = fmt.Errorf("Sending reply timed out after: '%v'.", tout)
		return
	}
	return
}

var lg *log.Logger

//
//  Implementation based on assumptions mentioned here:https://groups.google.com/forum/#!searchin/golang-nuts/TCPconn$20separate$20read$20write$20goroutine%7Csort:date/golang-nuts/EAm9FtsD_vk/kIhHulVjRn4J
//
func Start(
	cfg Config,
	lgr *log.Logger,
	wg *sync.WaitGroup,
	status chan<- string,
	term <-chan bool,
) (
	msgOut <-chan Msg,
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
	lg.Printf("Info: Listening for TCP connections on: '%s'.", addr.String())
	trig := make(chan bool)
	defer close(trig)
	msg := make(chan Msg)
	msgOut = msg
	wg.Add(1)
	go connManager(l, msg, trig, wg, status, term)
	// block till connManager ready to accept messages
	<-trig
	return
}

const deadLineInterval = 10 * time.Minute

func connManager(
	l *net.TCPListener,
	msg chan Msg,
	trig chan<- bool,
	wg *sync.WaitGroup,
	status chan<- string,
	term <-chan bool,
) {
	defer wg.Done()
	defer close(msg)
	lg.Printf("Info: Accepting TCP connections")
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
				lg.Printf("Retry: %d of %d\n", retryAttempts, retryMax)
				time.Sleep(time.Second * 2) //Gives it time to get its shit together
				goto connRetry
			}
			status <- "Failure"
			lg.Printf("Error: Failed after %d retries. Connection: '%s'.", retryAttempts, conn.RemoteAddr().String)
			status <- "Failure"
			return
		}
		if err := conn.SetDeadline(time.Now().Add(deadLineInterval)); err != nil {
			lg.Printf("Error: Deadline: '%s'.", err.Error())
			status <- "Failure"
			return
		}
		if err := conn.SetReadBuffer(PktSz); err != nil {
			lg.Printf("Error: Read Buffer: '%s'.", err.Error())
			status <- "Failure"
			return
		}
		if err := conn.SetWriteBuffer(PktSz); err != nil {
			lg.Printf("Error: Write Buffer: '%s'.", err.Error())
			status <- "Failure"
			return
		}
		lg.Printf("Info: Established client connection from: '%s'.", conn.RemoteAddr().String())
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
	abort := make(chan bool)
	defer close(abort)
	abortReply := make(chan bool)
	wg.Add(1)
	go closeConn(conn, wg, abort, abortReply, term)
	// failure to close channel intentional - closing channel will cause panic when terminating server
	reply := make(chan Msg)
	wg.Add(1)
	go replyMessage(conn, reply, wg, abortReply, term)

	for {
		m := NewMsg()
		if len, err := conn.Read(m.Pld); err != nil {
			if closed, _ := regexp.MatchString(".*closed network connection.*|.*EOF.*", err.Error()); closed {
				return
			}
			lg.Printf("Error: '%s' size of %d", err.Error(), len)
		}
		if err := conn.SetDeadline(time.Now().Add(deadLineInterval)); err != nil {
			lg.Printf("Error: Deadline: '%s'.", err.Error())
			return
		}
		lg.Printf("Info: Successfully read message '%s'", string(m.Pld))
		m.Addr = conn.RemoteAddr()
		m.reply = reply
		msg <- *m
	}
}
func replyMessage(conn *net.TCPConn, reply <-chan Msg, wg *sync.WaitGroup, abort chan bool, term <-chan bool) {
	defer wg.Done()
	defer close(abort)

	var m Msg
	for {
		select {
		case m = <-reply:
		case <-term:
			return
		}
		len, err := conn.Write(m.Pld)
		if err != nil {
			if closed, _ := regexp.MatchString(".*closed network connection.*|.*EOF.*", err.Error()); closed {
				return
			}
			lg.Printf("Error: While writing: '%s' size of: %d.", err.Error(), len)
		}
		if err := conn.SetDeadline(time.Now().Add(deadLineInterval)); err != nil {
			lg.Printf("Error: Deadline: '%s'.", err.Error())
			return
		}
		lg.Printf("Info: Completed sending reply message: '%s'.", string(m.Pld))
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
	abortRead <-chan bool,
	abortReply <-chan bool,
	term <-chan bool,
) {
	select {
	case <-abortRead:
	case <-abortReply:
	case <-term:
	}
	conn.Close()
	wg.Done()
}
