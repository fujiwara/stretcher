package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

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
		delay       float64
		sleep       time.Duration
	)
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Float64Var(&delay, "random-delay", 0, "sleep [0,random-delay) sec on start")
	flag.Parse()

	if showVersion {
		fmt.Println("version:", version)
		fmt.Println("build:", buildDate)
		checkLatest(version)
		return
	}
	log.Println("stretcher version:", version)

	if delay > 0 {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(delay*1000)))
		if err != nil {
			panic(err)
		}
		sleep, _ = time.ParseDuration(fmt.Sprintf("%dms", n.Int64()))
	}
	stretcher.Version = version
	stretcher.Init(sleep)
	err := stretcher.Run()
	if err != nil {
		log.Println(err)
		if os.Getenv("CONSUL_INDEX") != "" {
			// ensure exit 0 when running under `consul watch`
			return
		} else {
			os.Exit(1)
		}
	}
}

func fixVersionStr(v string) string {
	v = strings.TrimPrefix(v, "v")
	vs := strings.Split(v, "-")
	return vs[0]
}

func checkLatest(version string) {
	version = fixVersionStr(version)
	githubTag := &latest.GithubTag{
		Owner:             "fujiwara",
		Repository:        "stretcher",
		FixVersionStrFunc: fixVersionStr,
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
