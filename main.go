package main

import (
	"log"

	"github.com/2matzzz/mihari/mihari"
)

// func trace() (string, int, string) {
// 	pc, file, line, ok := runtime.Caller(1)
// 	if !ok {
// 		return "?", 0, "?"
// 	}

// 	fn := runtime.FuncForPC(pc)
// 	if fn == nil {
// 		return file, line, "?"
// 	}

// 	return file, line, fn.Name()
// }

// func init() {

// }

func main() {
	// conf := mihari.NewConfig()
	// var ports []Port
	port := &mihari.Port{
		Path: "/dev/ttyUSB3",
	}

	if err := port.Check(); err != nil {
		log.Printf("%#v", err)
	}
	log.Printf("%#v", port)
}
