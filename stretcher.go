package stretcher

import (
	"bufio"
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	LogBuffer bytes.Buffer
	Version   string
	s3svc     *s3.Client
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
	n := mrand.Int63n(int64(delay * Nanoseconds))
	return time.Duration(n)
}

func Run(ctx context.Context, conf Config) error {
	var err error
	log.Println("Starting up stretcher agent", Version)
	if conf.InitSleep > 0 {
		log.Printf("Sleeping %s", conf.InitSleep)
		tm := time.NewTimer(conf.InitSleep)
		select {
		case <-ctx.Done():
			// return immediately if context is canceled
			return ctx.Err()
		case <-tm.C:
		}
	}

	manifestURL, err := parseEvents(ctx)
	if err != nil {
		return fmt.Errorf("Could not parse event: %w", err)
	}

	log.Println("Loading manifest:", manifestURL)
	m, err := getManifest(ctx, manifestURL)
	if err != nil {
		return fmt.Errorf("Load manifest failed: %w", err)
	}
	log.Printf("Executing manifest %#v", m)

	err = m.Deploy(ctx, conf)
	if err != nil {
		log.Println("Deploy manifest failed:", err)
		if ferr := m.Commands.Failure.InvokePipe(ctx, &LogBuffer); ferr != nil {
			log.Println(ferr)
		}
		return fmt.Errorf("Deploy manifest failed: %s", err)
	}
	log.Println("Deploy manifest succeeded.")
	err = m.Commands.Success.InvokePipe(ctx, &LogBuffer)
	if err != nil {
		log.Println(err)
	}
	log.Println("Done")
	return nil
}

func getS3(ctx context.Context, u *url.URL) (io.ReadCloser, error) {
	svc, err := initS3()
	if err != nil {
		return nil, err
	}
	key := strings.TrimLeft(u.Path, "/")
	result, err := svc.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(u.Host),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func getGS(ctx context.Context, u *url.URL) (io.ReadCloser, error) {
	var err error
	trimPath := strings.Trim(u.Path, "/")

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

func getFile(_ context.Context, u *url.URL) (io.ReadCloser, error) {
	return os.Open(u.Path)
}

func getHTTP(ctx context.Context, u *url.URL) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
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

func getURL(ctx context.Context, urlStr string) (io.ReadCloser, error) {
	log.Println("Loading URL", urlStr)
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "s3":
		return getS3(ctx, u)
	case "gs":
		return getGS(ctx, u)
	case "http", "https":
		return getHTTP(ctx, u)
	case "file":
		return getFile(ctx, u)
	default:
		return nil, fmt.Errorf("manifest URL scheme must be s3, gs, http(s) or file: %s", urlStr)
	}
}

func getManifest(ctx context.Context, manifestURL string) (*Manifest, error) {
	rc, err := getURL(ctx, manifestURL)
	if err != nil {
		return nil, err
	}
	data, _ := io.ReadAll(rc)
	return ParseManifest(data)
}

func parseEvents(ctx context.Context) (string, error) {
	eventCh := make(chan string)
	var err error
	go func() {
		ev, e := parseEventsFromSTDIN()
		if e != nil {
			err = e
		}
		eventCh <- ev
	}()
	select {
	case <-ctx.Done():
		os.Stdin.Close()
		return "", ctx.Err()
	case ev := <-eventCh:
		return ev, err
	}
}

func parseEventsFromSTDIN() (string, error) {
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
		if err == io.EOF {
			// ignore EOF when the input has not LF.
			err = nil
		}
		return strings.Trim(line, "\n"), err
	}
}

func initS3() (*s3.Client, error) {
	if s3svc != nil {
		return s3svc, nil
	}
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	log.Println("region:", cfg.Region)
	return s3.NewFromConfig(cfg), nil
}
