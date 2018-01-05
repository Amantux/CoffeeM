package main


import (
	"strings"
	"io/ioutil"
	"fmt"
	"go.bug.st/serial.v1"	
	)
//Find The Arduino
func findarduino() string{
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		fmt.Println("Error0000:No Ports Found")
	return
	}
	contents, _ := ioutil.ReadDir("/dev")
	for _, f := range contents {
		if strings.Contains(f.Name(), "ttyUSB")|| strings.Contains(f.Name(), "tty.usbserial"){
			return "/dev/"+f.Name()
		}
	}
	return "Error0001:Arduino Connection not found"
}

//Open The Serial Channels
//Define Serial Comms
mode := &serial.Mode{
    BaudRate: 9600,
    Parity:   serial.NoParity,
    DataBits: 8,
    StopBits: serial.OneStopBit,
}

func sendArduino(toUUID string, []command string, fromUUID string, serialPort io.ReadWriteCloser) error{
	if serialPort == nil {
		return "Error0002:No Serial Found"
	}
	bufOut := new(bytes.Buffer)
	err := binary.Write(bufOut, binaryLittleEndian, argument)


func main(){
fmt.Println(findarduino())
}
