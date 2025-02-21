package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fujiwara/stretcher"
)

var (
	Version = "current"
)

func main() {
	var (
		showVersion  bool
		delay        float64
		maxBandWidth string
		timeout      int64
		retry        int
		retryWait    int64
		rsyncVerbose string
	)
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Float64Var(&delay, "random-delay", 0, "sleep [0,random-delay) sec on start")
	flag.StringVar(&maxBandWidth, "max-bandwidth", "", "max bandwidth for download src archives (Bytes/sec)")
	flag.Int64Var(&timeout, "timeout", 0, "timeout for download src archives (sec)")
	flag.IntVar(&retry, "retry", 0, "retry count for download src archives")
	flag.Int64Var(&retryWait, "retry-wait", 3, "wait for retry download src archives (sec)")
	flag.StringVar(&rsyncVerbose, "rsync-verbose", "-v", "rsync verbose revel default -v")
	flag.Parse()

	if showVersion {
		fmt.Println("version:", Version)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// backward compatibility before v0.10
	if region, found := os.LookupEnv("AWS_DEFAULT_REGION"); found && os.Getenv("AWS_REGION") == "" {
		os.Setenv("AWS_REGION", region)
	}
	os.Setenv("AWS_SDK_LOAD_CONFIG", "true")

	conf := stretcher.Config{
		InitSleep: stretcher.RandomTime(delay),
		Timeout:   time.Duration(timeout * int64(time.Second)),
		Retry:     retry,
		RetryWait: time.Duration(retryWait * int64(time.Second)),
	}
	if maxBandWidth != "" {
		if bw, err := humanize.ParseBytes(maxBandWidth); err != nil {
			fmt.Println("Cannot parse -max-bandwidth", err)
			os.Exit(1)
		} else {
			conf.MaxBandWidth = bw
		}
	}

	if err := stretcher.SetRsyncVerboseOpt(rsyncVerbose); err != nil {
		fmt.Println("Cannot set -rsync-verbose", err)
		os.Exit(1)
	}

	log.Println("stretcher version:", Version)
	stretcher.Version = Version
	err := stretcher.Run(ctx, conf)
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
