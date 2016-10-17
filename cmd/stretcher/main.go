package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fujiwara/stretcher"
	"github.com/tcnksm/go-latest"
)

var (
	version   string
	buildDate string
)

func main() {
	var (
		showVersion  bool
		delay        float64
		maxBandWidth string
		timeout      int64
		retry        int
	)
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Float64Var(&delay, "random-delay", 0, "sleep [0,random-delay) sec on start")
	flag.StringVar(&maxBandWidth, "max-bandwidth", "", "max bandwidth for download src archives (Bytes/sec)")
	flag.Int64Var(&timeout, "timeout", 0, "timeout for download src archives (sec)")
	flag.IntVar(&retry, "retry", 0, "retry count for download src archives")
	flag.Parse()

	if showVersion {
		fmt.Println("version:", version)
		fmt.Println("build:", buildDate)
		checkLatest(version)
		return
	}

	conf := stretcher.Config{
		InitSleep: stretcher.RandomTime(delay),
		Timeout:   time.Duration(timeout * int64(time.Second)),
		Retry:     retry,
	}
	if maxBandWidth != "" {
		if bw, err := humanize.ParseBytes(maxBandWidth); err != nil {
			fmt.Println("Cannot parse -max-bandwidth", err)
			os.Exit(1)
		} else {
			conf.MaxBandWidth = bw
		}
	}

	log.Println("stretcher version:", version)
	stretcher.Version = version
	err := stretcher.Run(conf)
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
