package stretcher

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

type Commands struct {
	Pre     CommandLines `yaml:"pre"`
	Post    CommandLines `yaml:"post"`
	Success CommandLines `yaml:"success"`
	Failure CommandLines `yaml:"failure"`
}

type CommandLines []CommandLine

func (cs CommandLines) Invoke() error {
	for _, c := range cs {
		err := c.Invoke()
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs CommandLines) InvokePipe(src *bytes.Buffer) error {
	for _, c := range cs {
		buf := bytes.NewBuffer(src.Bytes())
		err := c.InvokePipe(buf)
		if err != nil {
			return err
		}
	}
	return nil
}

type CommandLine string

func (c CommandLine) String() string {
	return string(c)
}

func (c CommandLine) Invoke() error {
	log.Println("invoking command:", c.String())
	out, err := exec.Command("sh", "-c", c.String()).CombinedOutput()
	if len(out) > 0 {
		log.Println(string(out))
	}
	if err != nil {
		return fmt.Errorf("failed: %v %v", c, err)
	}
	return nil
}

func (c CommandLine) InvokePipe(src io.Reader) error {
	log.Println("invoking command:", c.String())
	cmd := exec.Command("sh", "-c", c.String())
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(3)
	// src => cmd.stdin
	go func() {
		_, err := io.Copy(stdin, src)
		if err != nil {
			log.Println(err)
		}
		stdin.Close()
		wg.Done()
	}()
	// cmd.stdout => stretcher.stdout
	go func() {
		_, err := io.Copy(os.Stdout, stdout)
		if err != nil {
			log.Println(err)
		}
		stdout.Close()
		wg.Done()
	}()
	// cmd.stderr => stretcher.stderr
	go func() {
		_, err := io.Copy(os.Stderr, stderr)
		if err != nil {
			log.Println(err)
		}
		stderr.Close()
		wg.Done()
	}()
	wg.Wait()
	return cmd.Wait()
}
