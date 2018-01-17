package main


import (
	"strings"
	"io/ioutil"
	"fmt"
	"github.com/huin/goserial"
	"log"
	)
//Find The Arduino
func findarduino() string{
	contents, _ := ioutil.ReadDir("/dev")
	for _, f := range contents {
		if strings.Contains(f.Name(), "ttyUSB")|| strings.Contains(f.Name(), "tty.usbserial"){
			return "/dev/"+f.Name()
		}
	}
	return "Error0001:Arduino Connection not found"
}

//Serial Config Session
func OpenSerial(port string, SerialRate int){
        c := &serial.Config{Name: port, Baud: SerialRate}
        s, err := serial.OpenPort(c)
        if err != nil {
                log.Fatal(err)
	}
	return s;
}
//Opens Serial, returns open port


//Open Serial Ports with Prior SettingS

func OpenPort(Serial int){
	port, err := serial.Open(findarduino(), mode)
	if err != nil {
    log.Fatal(err)
	}
}

//Format Message
func FormatMessage(message string){
	return []byte(message)
}
func main(){
fmt.Println(findarduino())
}
