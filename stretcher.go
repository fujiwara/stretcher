package stretcher

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/s3"
)

var (
	AWSAuth   aws.Auth
	AWSRegion aws.Region
	LogBuffer bytes.Buffer
)

func Init() {
	log.SetOutput(io.MultiWriter(os.Stderr, &LogBuffer))
}

func Run() error {
	log.Println("Starting up stretcher agent")

	var err error
	payload, err := parseEvents()
	if err != nil {
		return fmt.Errorf("Could not parse event: %s", err)
	}

	log.Println("Loading manifest:", payload)
	m, err := getManifest(payload)
	if err != nil {
		return fmt.Errorf("Load manifest failed: %s", err)
	}
	log.Printf("Executing manifest %#v", m)

	err = m.Deploy()
	if err != nil {
		log.Println("Deploy manifest failed:", err)
		m.Commands.Failure.InvokePipe(&LogBuffer)
		return fmt.Errorf("Deploy manifest failed: %s", err)
	}
	log.Println("Deploy manifest succeeded.")
	m.Commands.Success.InvokePipe(&LogBuffer)
	return nil
}

func getS3(u *url.URL) (io.ReadCloser, error) {
	var err error
	if AWSAuth.AccessKey == "" || AWSRegion.Name == "" {
		AWSAuth, AWSRegion, err = LoadAWSCredentials("")
		if err != nil {
			return nil, err
		}
		log.Println("region:", AWSRegion.Name)
		log.Println("aws_access_key_id:", AWSAuth.AccessKey)
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
	log.Println("Loading URL", urlStr)
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

func parseEvents() (string, error) {
	log.Println("Waiting for events from STDIN...")
	if os.Getenv("CONSUL_INDEX") != "" {
		log.Println("Reading Consul event")
		ev, err := ParseConsulEvents(os.Stdin)
		if err != nil {
			return "", err
		}
		if ev == nil {
			// no event
			return "", fmt.Errorf("No Consul events found")
		}
		return ev.PayloadString(), nil
	} else {
		if userEvent := os.Getenv("SERF_USER_EVENT"); userEvent != "" {
			log.Println("Reading Serf user event:", userEvent)
		}
		// event passed by stdin (raw string)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			return scanner.Text(), nil
		}
		return "", scanner.Err()
	}
}
