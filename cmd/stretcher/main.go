package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/fujiwara/stretcher"
	"github.com/tcnksm/go-latest"
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
		checkLatest(version)
		return
	}
	log.Println("stretcher version:", version)
	stretcher.Init()
	stretcher.Run()
}

func checkLatest(version string) {
	githubTag := &latest.GithubTag{
		Owner:      "fujiwara",
		Repository: "stretcher",
	}
	res, err := latest.Check(githubTag, version)
	if err != nil {
		fmt.Println(err)
		return
	}
	if res.Outdated {
		fmt.Printf("%s is not latest, you should upgrade to %s\n", version, res.Current)
	}
}
