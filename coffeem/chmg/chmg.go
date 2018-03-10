package chmg

import (
	"CoffeeM/coffeem/udp"
	"log"
	"sync"
)

var lg *log.Logger

func Start(destinationChan chan<- udp.Msg, sourceChan <-chan udp.Msg, lgr *log.Logger) {
	var lockLgr sync.Mutex
	lockLgr.Lock()
	if lg == nil && lgr != nil {
		lg = lgr
	}
	lockLgr.Unlock()
	go pass(destinationChan, sourceChan)

}
func pass(destinationChan chan<- udp.Msg, sourceChan <-chan udp.Msg) {
	for {
		select {
		case msg, alive := <-sourceChan:
			if !alive {
				return
			}
			select {
			case destinationChan <- msg:
			default:
				lg.Println("chmg:DestinationChan Full")
			}
		}
	}
}
