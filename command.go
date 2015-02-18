package stretcher

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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
	if err != nil {
		return fmt.Errorf("failed: %v %v", c, err)
	}
	log.Println(string(out))
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
	cmdCh := make(chan error)
	// src => stdin
	go func() {
		_, err := io.Copy(stdin, src)
		if err != nil {
			cmdCh <- err
		}
		stdin.Close()
	}()
	// wait for command exit
	go func() {
		cmdCh <- cmd.Wait()
	}()
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	cmdErr := <-cmdCh
	return cmdErr
}
