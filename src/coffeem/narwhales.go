package main


import (
	"strings"
	"io/ioutil"
	"fmt"
	"github.com/huin/goserial"
	)

func findarduino() string{
	contents, _ := ioutil.ReadDir("/dev")
	for _, f := range contents {
		if strings.Contains(f.Name(), "ttyUSB")|| strings.Contains(f.Name(), "tty.usbserial"){
			return "/dev/"+f.Name()
		}
	}
	return "Error0001:Arduino Connection not found"
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
