package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/fujiwara/stretcher"
)

func main() {
	var profile string
	flag.StringVar(&profile, "profile", "", "aws profile name")
	flag.Parse()

	auth, region, err := stretcher.LoadAWSCredentials(profile)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("profile     ", profile)
	fmt.Println("access_key  ", auth.AccessKey)
	fmt.Println("secret_key  ", auth.SecretKey)
	fmt.Println("region      ", region.Name)
}
