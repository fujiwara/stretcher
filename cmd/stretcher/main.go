package main

import (
	"flag"
	"fmt"
	"github.com/fujiwara/stretcher"
	"log"
)

var (
	version   string
	buildDate string
)

func main() {
	var (
		showVersion bool
	)
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()

	if showVersion {
		fmt.Println("version:", version)
		fmt.Println("build:", buildDate)
		return
	}
	log.Println("stretcher version:", version)
	stretcher.Run()
}
