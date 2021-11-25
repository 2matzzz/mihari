package main

import (
	"fmt"
	"log"

	"go.bug.st/serial"
)

func main() {
	port, err := serial.Open("/dev/ttyACM0", &serial.Mode{})
	if err != nil {
		log.Fatal(err)
	}
	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	if err := port.SetMode(mode); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Port set to 115200 N81")
}
