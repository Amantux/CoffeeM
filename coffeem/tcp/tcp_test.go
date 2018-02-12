package tcp

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func xTest_Startup(t *testing.T) {
	cfg := Config{
		IPListen: "127.0.0.1",
		Port:     10001,
	}
	msg := make(chan Msg)
	defer close(msg)
	lg := log.New(os.Stderr, "Pi:", log.LstdFlags)
	term := make(chan bool)
	go termRasPI(term)
	status := make(chan string)
	defer close(status)
	lg.Print("hi")
	var wg sync.WaitGroup
	Start(cfg, msg, lg, &wg, status, term)

	// coordinate graceful terination
	select {
	case <-term:
	case <-status:
		// suiside call
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		<-term
	}
	lg.Print("before wait")
	wg.Wait()
}

func Test_OneMsg(t *testing.T) {
	cfg := Config{
		IPListen: "127.0.0.1",
		Port:     10001,
	}
	msg := make(chan Msg)
	lg := log.New(os.Stderr, "Pi:", log.LstdFlags)
	term := make(chan bool)
	go termRasPI(term)
	status := make(chan string)
	defer close(status)
	lg.Print("hi")
	var wg sync.WaitGroup
	Start(cfg, msg, lg, &wg, status, term)
	lg.Print("after start")
	go msgPrint(msg, &wg)
	msgClient := make(chan Msg)
	go sendMsg(t, cfg, msgClient)
	genMsg(t, 1, msgClient)
	close(msgClient)
	lg.Print("after gen message")

	// coordinate graceful terination
	select {
	case <-term:
	case <-status:
		// suiside call
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		<-term
	}
	lg.Print("before wait")
	wg.Wait()
	wg.Add(1)
	close(msg)
	wg.Wait()

}

func genMsg(t *testing.T, tot uint16, msg chan<- Msg) {
	t.Logf("Generating messages. Total: %d\n", tot)
	for c := uint16(1); c < tot+1; c++ {
		m := NewMsg()
		m.Pld = []byte(fmt.Sprintf("Msg Payload: %d", c))
		msg <- *m
	}
}

func sendMsg(t *testing.T, cfg Config, msg <-chan Msg) {
	// resolve address to server
	t.Logf("Resolve Server Address.\n")
	tcpendpt := cfg.IPListen
	tcpendpt += ":" + strconv.Itoa(cfg.Port)
	ServerAddr, errServ := net.ResolveTCPAddr("tcp", tcpendpt)
	assert.NoError(t, errServ)
	// resolve address of this client
	tcpendpt = "localhost"
	tcpendpt += ":" + strconv.Itoa(cfg.Port+1)
	t.Logf("Resolve Client Address.\n")
	LocalAddr, errLcl := net.ResolveTCPAddr("tcp", tcpendpt)
	assert.NoError(t, errLcl)
	// open connection between this cliend and its server
	t.Logf("Dialing Server.\n")
	Conn, errConn := net.DialTCP("tcp", LocalAddr, ServerAddr)
	if errConn != nil {
		t.Fatalf("Error:making connection\n")
	}
	assert.NoError(t, errConn)
	defer Conn.Close()
	// send messages over the connection
	t.Logf("Sending messages.\n")
	msgTot := 0
	for m := range msg {
		_, err := Conn.Write(m.Pld)
		assert.NoError(t, err)
		msgTot++
		fmt.Fprintf(os.Stderr, "in Send Msg \n")
	}
	t.Logf("Sent messages. Total: %d\n", msgTot)
}

func msgPrint(msg <-chan Msg, wg *sync.WaitGroup) {
	for m := range msg {
		fmt.Println(m.Pld)
	}
	wg.Done()
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
}
