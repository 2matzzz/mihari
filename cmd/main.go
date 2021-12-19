package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/2matzzz/mihari"
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

	//TODO: logger

	// logger := &log.Logger{}
	config, err := mihari.NewConfig("example/mihari.yml")
	if err != nil {
		log.Fatalln(err)
		// logger.Fatalln(err)
	}
	client, err := mihari.NewClient(config)
	if err != nil {
		log.Fatalln(err)
		// logger.Fatalln(err)
	}
	defer client.Close()

	if err := client.Check(); err != nil {
		log.Fatalln(err)
		// logger.Fatalln(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		client.Close()
		os.Exit(1)
	}()
	// client.Run()
}
