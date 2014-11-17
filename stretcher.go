package stretcher

import (
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var (
	AWSAuth   aws.Auth
	AWSRegion aws.Region
)

func Run() {
	log.Println("Starting up stretcher agent")

	if file := os.Getenv("AWS_CONFIG_FILE"); file != "" {
		err := LoadAWSConfigFile(file)
		if err != nil {
			log.Println("Load AWS_CONFIG_FILE failed:", err)
			return
		}
	}

	log.Println("Waiting for consul events from STDIN...")

	ev, err := ParseConsulEvents(os.Stdin)
	if err != nil {
		log.Println("Parse consul events failed:", err)
		return
	}
	if ev == nil {
		// no event
		return
	}

	log.Println("Executing manifest:", ev.PayloadString())
	manifest, err := getManifest(ev.PayloadString())
	if err != nil {
		log.Println("Load manifest failed:", err)
		return
	}
	log.Printf("%#v", manifest)

	err = manifest.Deploy()
	if err != nil {
		log.Println("Deploy manifest failed:", err)
	}
}

func getS3(u *url.URL) (io.ReadCloser, error) {
	if AWSAuth.AccessKey == "" || AWSRegion.Name == "" {
		return nil, fmt.Errorf("Invalid AWS Auth or Region. Please check env AWS_CONFIG_FILE.")
	}
	client := s3.New(AWSAuth, AWSRegion)
	bucket := client.Bucket(u.Host)
	rc, err := bucket.GetReader(u.Path)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func getFile(u *url.URL) (io.ReadCloser, error) {
	file, err := os.Open(u.Path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func getHTTP(u *url.URL) (io.ReadCloser, error) {
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func getURL(urlStr string) (io.ReadCloser, error) {
	log.Println("loading URL", urlStr)
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "s3":
		return getS3(u)
	case "http", "https":
		return getHTTP(u)
	case "file":
		return getFile(u)
	default:
		return nil, fmt.Errorf("manifest URL scheme must be s3 or http(s) or file: %s", urlStr)
	}
}

func getManifest(manifestURL string) (*Manifest, error) {
	rc, err := getURL(manifestURL)
	if err != nil {
		return nil, err
	}
	data, _ := ioutil.ReadAll(rc)
	return ParseManifest(data)
}
