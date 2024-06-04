package stretcher

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

type Commands struct {
	Pre     CommandLines `yaml:"pre"`
	Post    CommandLines `yaml:"post"`
	Success CommandLines `yaml:"success"`
	Failure CommandLines `yaml:"failure"`
}

type CommandLines []CommandLine

func (cs CommandLines) Invoke(ctx context.Context) error {
	for _, c := range cs {
		err := c.Invoke(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs CommandLines) InvokePipe(ctx context.Context, src *bytes.Buffer) error {
	for _, c := range cs {
		buf := bytes.NewBuffer(src.Bytes())
		err := c.InvokePipe(ctx, buf)
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

func (c CommandLine) Invoke(ctx context.Context) error {
	log.Println("invoking command:", c.String())
	out, err := exec.CommandContext(ctx, "sh", "-c", c.String()).CombinedOutput()
	if len(out) > 0 {
		log.Println(string(out))
	}
	if err != nil {
		return fmt.Errorf("failed: %v %v", c, err)
	}
	return nil
}

func (c CommandLine) InvokePipe(ctx context.Context, src io.Reader) error {
	log.Println("invoking command:", c.String())
	cmd := exec.CommandContext(ctx, "sh", "-c", c.String())
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
		if e, ok := err.(*os.PathError); ok && e.Err == syscall.EPIPE {
			// ignore EPIPE
		} else if err != nil {
			log.Println("failed to write to STDIN of", c.String(), ":", err)
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
