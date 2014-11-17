package main

import (
	"flag"
	"fmt"
	"github.com/fujiwara/stretcher"
	"os"
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
		os.Exit(0)
	}
	stretcher.Run()
}
