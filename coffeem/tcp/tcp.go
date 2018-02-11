package tcp

import (
	"log"
	"net"
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
	status <-chan string,
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
	trig := make(chan bool)
	defer close(trig)
	wg.Add(1)
	go connManager(c, msg, trig, wg, status, term)
	// block till connManager ready to accept messages
	<-trig
	return
}
func connManager(
	l *net.TCPListener,
	msg chan<- Msg,
	trig chan<- bool,
	wg *sync.WaitGroup,
	status <-chan string,
	term <-chan bool,
) {
	defer wg.Done()
	defer l.Close()
	// signal connection manager started
	trig <- true

	for {
		retryAttmpts := 0
	connRetry:
		retryAttempts += 1
		conn, err := l.Accept()
		if err != nil {
			lg.Printf("Error Accepting: %s \n", err.Error())
			if retryAttmpts < 5 {
				lg.Printf("Retry Attmpt: %d of %d\n", retryAttempts, 5)
				time.Sleep(time.Second * 2) //Gives it time to get its shit together
				goto connRetry
			}
			status
			lg.Printf("Error failed after %d retries. Connection: %v\n", retryAttmpts, conn)
			status <- "Failure"
			return
		}
		retryAttmpts := 0
		timeoutDuration := 10 * time.Minute
		if err := conn.SetReadDeadline(time.Now().Add(timeoutDuration)); err != nil {
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
		go readMessage(c, msg, wg, term)
	}
}
func readMessage(
	conn *net.Conn,
	msg chan<- Msg,
	wg *sync.WaitGroup,
	term <-chan bool,
) {
	defer wg.Done()
	wg.Add(1)
	go closeConn(conn, term)
	for {
		m := NewMsg()
		if len, err := conn.Read(m.Pld); err != nil || len != PktSz {
			lg.Printf("Error Reading: %s size of %d \n", err.Error(), len)
		}
		if err := conn.SetReadDeadline(time.Now().Add(timeoutDuration)); err != nil {
			lg.Printf("Error Read Deadline: %s \n", err.Error())
			return
		}
	}
}
func closeConn(
	conn *net.Conn,
	wg *sync.WaitGroup,
	term <-chan bool,
) {
	<-term
	conn.close()
	wg.Done()
}
