package stretcher

import (
	"bufio"
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/s3"
)

var (
	AWSAuth   aws.Auth
	AWSRegion aws.Region
	LogBuffer bytes.Buffer
	Version   string
)

type Config struct {
	MaxBandWidth uint64
	Timeout      time.Duration
	InitSleep    time.Duration
	Retry        int
	RetryWait    time.Duration
}

const Nanoseconds = 1000 * 1000 * 1000

func init() {
	log.SetOutput(io.MultiWriter(os.Stderr, &LogBuffer))
}

func RandomTime(delay float64) time.Duration {
	if delay <= 0 {
		return time.Duration(0)
	}
	// http://stackoverflow.com/questions/6181260/generating-random-numbers-in-go
	var s int64
	if err := binary.Read(crand.Reader, binary.LittleEndian, &s); err != nil {
		s = time.Now().UnixNano()
	}
	mrand.Seed(s)
	n := mrand.Int63n(int64(delay * Nanoseconds))
	return time.Duration(n)
}

func Run(conf Config) error {
	var err error
	log.Println("Starting up stretcher agent", Version)
	if conf.InitSleep > 0 {
		log.Printf("Sleeping %s", conf.InitSleep)
		time.Sleep(conf.InitSleep)
	}

	manifestURL, err := parseEvents()
	if err != nil {
		return fmt.Errorf("Could not parse event: %s", err)
	}

	log.Println("Loading manifest:", manifestURL)
	m, err := getManifest(manifestURL)
	if err != nil {
		return fmt.Errorf("Load manifest failed: %s", err)
	}
	log.Printf("Executing manifest %#v", m)

	err = m.Deploy(conf)
	if err != nil {
		log.Println("Deploy manifest failed:", err)
		if ferr := m.Commands.Failure.InvokePipe(&LogBuffer); ferr != nil {
			log.Println(ferr)
		}
		return fmt.Errorf("Deploy manifest failed: %s", err)
	}
	log.Println("Deploy manifest succeeded.")
	err = m.Commands.Success.InvokePipe(&LogBuffer)
	if err != nil {
		log.Println(err)
	}
	log.Println("Done")
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
	return bucket.GetReader(u.Path)
}

func getGS(u *url.URL) (io.ReadCloser, error) {
	var err error
	trimPath := strings.Trim(u.Path, "/")

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(u.Host).Object(trimPath).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}

func getFile(u *url.URL) (io.ReadCloser, error) {
	return os.Open(u.Path)
}

func getHTTP(u *url.URL) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Stretcher/"+Version)

	resp, err := http.DefaultClient.Do(req)
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
	case "gs":
		return getGS(u)
	case "http", "https":
		return getHTTP(u)
	case "file":
		return getFile(u)
	default:
		return nil, fmt.Errorf("manifest URL scheme must be s3, gs, http(s) or file: %s", urlStr)
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
	reader := bufio.NewReader(os.Stdin)
	b, err := reader.Peek(1)
	if err != nil {
		return "", err
	}
	if os.Getenv("CONSUL_INDEX") != "" || string(b) == "[" {
		log.Println("Reading Consul event")
		ev, err := ParseConsulEvents(reader)
		if err != nil {
			return "", err
		}
		return string(ev.Payload), nil
	} else {
		if userEvent := os.Getenv("SERF_USER_EVENT"); userEvent != "" {
			log.Println("Reading Serf user event:", userEvent)
		}
		// event passed by stdin (raw string)
		line, err := reader.ReadString('\n')
		return strings.Trim(line, "\n"), err
	}
}
