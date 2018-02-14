package tcp

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Startup(t *testing.T) {
	cfg := Config{
		IPListen: "127.0.0.1",
		Port:     randPort(),
	}
	lg := log.New(os.Stderr, "Pi:", log.LstdFlags)
	term := make(chan bool)
	go termRasPI(term)
	status := make(chan string)
	defer close(status)
	var wg sync.WaitGroup
	if _, err := Start(cfg, lg, &wg, status, term); err != nil {
		lg.Printf("Error: When starting TCP server: '%s'.", err.Error())
		t.Fail()
	}
	// coordinate graceful terination
	select {
	case <-term:
	case <-status:
		// suicide call
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		<-term
	}
	lg.Print("Info: Before Wait")
	wg.Wait()
}

func Test_OneMsg(t *testing.T) {
	cfg := Config{
		IPListen: "127.0.0.1",
		Port:     randPort(),
	}
	lg := log.New(os.Stderr, "Pi:", log.LstdFlags)
	term := make(chan bool)
	go termRasPI(term)
	status := make(chan string)
	defer close(status)
	lg.Print("Starting Server")
	var wgServer sync.WaitGroup
	msg, err := Start(cfg, lg, &wgServer, status, term)
	if err != nil {
		lg.Printf("Error: When starting TCP server: '%s'.", err.Error())
		t.Fail()
	}
	lg.Print("Server Up!")
	var wgServerOut sync.WaitGroup
	wgServerOut.Add(1)
	go msgPrint(t, msg, &wgServerOut)
	// start the client
	msgClient := make(chan Msg)
	var wgClient sync.WaitGroup
	wgClient.Add(1)
	go sendMsg(t, cfg, msgClient, &wgClient)
	lg.Print("Client: Generating Messages")
	genMsg(t, 1, msgClient)
	close(msgClient)
	wgClient.Wait()
	lg.Print("Client Complete")
	// coordinate graceful terination
	select {
	case <-term:
	case <-status:
		// suicide call
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		<-term
	}
	lg.Print("Server Terminated - waiting until it quiesces.")
	wgServer.Wait()
	lg.Print("Server Stopped - force out remaining messages.")
	wgServerOut.Wait()
}

func XTest_TwoMsg(t *testing.T) {
	cfg := Config{
		IPListen: "127.0.0.1",
		Port:     randPort(),
	}
	lg := log.New(os.Stderr, "Pi:", log.LstdFlags)
	term := make(chan bool)
	go termRasPI(term)
	status := make(chan string)
	defer close(status)
	lg.Print("Starting Server")
	var wgServer sync.WaitGroup
	msg, err := Start(cfg, lg, &wgServer, status, term)
	if err != nil {
		lg.Printf("Error: When starting TCP server: '%s'.", err.Error())
		t.Fail()
	}
	lg.Print("Server Up!")
	var wgServerOut sync.WaitGroup
	wgServerOut.Add(1)
	go msgPrint(t, msg, &wgServerOut)
	// start the client
	msgClient := make(chan Msg)
	var wgClient sync.WaitGroup
	wgClient.Add(1)
	go sendMsg(t, cfg, msgClient, &wgClient)
	lg.Print("Client: Generating Messages")
	genMsg(t, 2, msgClient)
	close(msgClient)
	wgClient.Wait()
	lg.Print("Client Complete")
	// coordinate graceful terination
	select {
	case <-term:
	case <-status:
		// suicide call
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		<-term
	}
	lg.Print("Server Terminated - waiting until it quiesces.")
	wgServer.Wait()
	lg.Print("Server Stopped - force out remaining messages.")
	wgServerOut.Wait()
}

func Test_1000Msg(t *testing.T) {
	cfg := Config{
		IPListen: "127.0.0.1",
		Port:     randPort(),
	}
	lg := log.New(os.Stderr, "Pi:", log.LstdFlags)
	term := make(chan bool)
	go termRasPI(term)
	status := make(chan string)
	defer close(status)
	lg.Print("Starting Server")
	var wgServer sync.WaitGroup
	msg, err := Start(cfg, lg, &wgServer, status, term)
	if err != nil {
		lg.Printf("Error: When starting TCP server: '%s'.", err.Error())
		t.Fail()
	}
	lg.Print("Server Up!")
	var wgServerOut sync.WaitGroup
	wgServerOut.Add(1)
	go msgPrint(t, msg, &wgServerOut)
	// start the client
	msgClient := make(chan Msg)
	var wgClient sync.WaitGroup
	wgClient.Add(1)
	go sendMsg(t, cfg, msgClient, &wgClient)
	lg.Print("Client: Generating Messages")
	genMsg(t, 1000, msgClient)
	close(msgClient)
	wgClient.Wait()
	lg.Print("Client Complete")
	// coordinate graceful terination
	select {
	case <-term:
	case <-status:
		// suicide call
		close(term)
	}
	lg.Print("Server Terminated - waiting until it quiesces.")
	wgServer.Wait()
	lg.Print("Server Stopped - force out remaining messages.")
	wgServerOut.Wait()
}
func genMsg(t *testing.T, tot uint16, msg chan<- Msg) {
	t.Logf("Generating messages. Total: %d\n", tot)
	for c := uint16(1); c < tot+1; c++ {
		m := NewMsg()
		m.Pld = []byte(fmt.Sprintf("Msg Payload: %d", c))
		msg <- *m
	}
}

func sendMsg(t *testing.T, cfg Config, msg <-chan Msg, wg *sync.WaitGroup) {
	defer wg.Done()
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
	t.Logf("Dialing server: '%s' client: '%s'\n", ServerAddr.String(), LocalAddr.String())
	Conn, errConn := net.DialTCP("tcp", LocalAddr, ServerAddr)
	if errConn != nil {
		t.Fatalf("Error: making connection\n")
	}
	assert.NoError(t, errConn)
	defer Conn.Close()
	assert.NoError(t, Conn.SetReadBuffer(PktSz))
	assert.NoError(t, Conn.SetWriteBuffer(PktSz))
	// send messages over the connection
	t.Logf("Sending messages.\n")
	msgTot := 0
	for m := range msg {
		_, err := Conn.Write(m.Pld)
		assert.NoError(t, err)
		msgTot++
		t.Logf("Info: Sending message: '%d'. ", msgTot)
	}
	t.Logf("Info: Sent messages. Total: %d\n", msgTot)
}

func msgPrint(t *testing.T, msg <-chan Msg, wg *sync.WaitGroup) {
	for m := range msg {
		t.Logf("Info: message out channel pushed message: '%s'.", string(m.Pld))
	}
	wg.Done()
}
func randPort() int {
	daytime := time.Now()
	dayHr, dayMin, DaySec := daytime.Clock()
	dayAgeNow := dayHr*3600 + dayMin*60 + DaySec
	src := rand.NewSource(int64(dayAgeNow))
	ranGen := rand.New(src)
	portOffset := ranGen.Intn(54000)
	return 1000 + portOffset
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
