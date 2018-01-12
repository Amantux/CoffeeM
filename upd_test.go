package udp

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	ucfg := Config{
		IPListen: "0.0.0.0",
		Port:     3330,
	}
	msg := make(chan Msg)
	lgr := log.New(os.Stderr, "RasPi:  ", log.LstdFlags)
	term := make(chan bool)
	var wg sync.WaitGroup
	assert.NoError(t, Start(ucfg, msg, lgr, &wg, term))
	wg.Add(1)
	go msgPrint(msg, &wg)
	close(term)
	close(msg)
	wg.Wait()
}

func TestStartOneMess(t *testing.T) {
	ucfg := Config{
		IPListen: "0.0.0.0",
		Port:     3330,
	}
	msg := make(chan Msg)
	lgr := log.New(os.Stderr, "RasPi:  ", log.LstdFlags)
	term := make(chan bool)
	var wg sync.WaitGroup
	assert.NoError(t, Start(ucfg, msg, lgr, &wg, term))
	wg.Add(1)
	go msgPrint(msg, &wg)
	lgr.Println("resolving server")
	ServerAddr, errServ := net.ResolveUDPAddr("udp4", "127.0.0.1:3330")
	assert.NoError(t, errServ)
	lgr.Println("resolving local")
	LocalAddr, errLcl := net.ResolveUDPAddr("udp4", "127.0.30.1:3033")
	assert.NoError(t, errLcl)
	lgr.Println("dialing server")
	Conn, errConn := net.DialUDP("udp4", LocalAddr, ServerAddr)
	assert.NoError(t, errConn)
	defer Conn.Close()
	for i := 0; i < 10; i++ {
		msgPld := strconv.Itoa(i)
		buf := []byte(msgPld)
		_, err := Conn.Write(buf)
		assert.NoError(t, err)
	}
	time.Sleep(time.Millisecond * 1) //Close runs real fast, add in delay for message to be sent out, add in ACK to provide Parity

	close(term)
	close(msg)
	wg.Wait()
}

func msgPrint(msg <-chan Msg, wg *sync.WaitGroup) {
	for m := range msg {
		fmt.Println(m.Pld)
	}
	wg.Done()
}

func blockDur(t time.Duration) func() {
	return func() {
		time.Sleep(t)
	}
}

func termSignal(nodeCnt int, term chan bool, block func()) {
	block()
	for i := nodeCnt; i > 0; i-- {
		term <- true
	}
}
