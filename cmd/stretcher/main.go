package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/fujiwara/stretcher"
)

var (
	Version = "current"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("stretcher version:", Version)
	stretcher.Version = Version

	conf := &stretcher.Config{}
	kong.Parse(conf)
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
