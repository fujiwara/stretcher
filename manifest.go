package stretcher

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"gopkg.in/yaml.v2"
)

var RsyncDefaultOpts = []string{"-av", "--delete"}
var DefaultDestMode = os.FileMode(0755)

type Manifest struct {
	Src         string       `yaml:"src"`
	CheckSum    string       `yaml:"checksum"`
	Dest        string       `yaml:"dest"`
	DestMode    *os.FileMode `yaml:"dest_mode"`
	Commands    Commands     `yaml:"commands"`
	Excludes    []string     `yaml:"excludes"`
	ExcludeFrom string       `yaml:"exclude_from"`
}

func (m *Manifest) newHash() (hash.Hash, error) {
	switch len(m.CheckSum) {
	case 32:
		return md5.New(), nil
	case 40:
		return sha1.New(), nil
	case 64:
		return sha256.New(), nil
	case 128:
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("checksum must be md5, sha1, sha256, sha512 hex string.")
	}
}

func (m *Manifest) Deploy() error {
	begin := time.Now()
	src, err := getURL(m.Src)
	if err != nil {
		return fmt.Errorf("Get src failed:", err)
	}
	defer src.Close()

	tmp, err := ioutil.TempFile(os.TempDir(), "stretcher")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	written, sum, err := m.copyAndCalcHash(tmp, src)
	tmp.Close()
	if err != nil {
		return err
	}
	duration := float64(time.Now().Sub(begin).Nanoseconds()) / 1000000000
	log.Printf("Wrote %s bytes to %s (in %s sec, %s/s)",
		humanize.Comma(written),
		tmp.Name(),
		humanize.Ftoa(duration),
		humanize.Bytes(uint64(float64(written)/duration)),
	)
	if len(m.CheckSum) > 0 && sum != strings.ToLower(m.CheckSum) {
		return fmt.Errorf("Checksum mismatch. expected:%s got:%s", m.CheckSum, sum)
	} else {
		log.Printf("Checksum ok: %s", sum)
	}

	dir, err := ioutil.TempDir(os.TempDir(), "stretcher_src")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	err = m.Commands.Pre.Invoke()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err = os.Chdir(dir); err != nil {
		return err
	}

	log.Println("Extract archive:", tmp.Name(), "to", dir)
	out, err := exec.Command("tar", "xf", tmp.Name()).CombinedOutput()
	if len(out) > 0 {
		log.Println(string(out))
	}
	if err != nil {
		log.Println("failed: tar xf", tmp.Name(), "failed", err)
		return err
	}

	log.Println("Set dest mode", *m.DestMode)
	err = os.Chmod(dir, *m.DestMode)
	if err != nil {
		return err
	}

	from := dir + "/"
	to := m.Dest
	// append "/" when not terminated by "/"
	if strings.LastIndex(to, "/") != len(to)-1 {
		to = to + "/"
	}

	args := []string{}
	args = append(args, RsyncDefaultOpts...)
	if m.ExcludeFrom != "" {
		args = append(args, "--exclude-from", from+m.ExcludeFrom)
	}
	if len(m.Excludes) > 0 {
		for _, ex := range m.Excludes {
			args = append(args, "--exclude", ex)
		}
	}
	args = append(args, from, to)

	log.Println("rsync", args)
	out, err = exec.Command("rsync", args...).CombinedOutput()
	if len(out) > 0 {
		log.Println(string(out))
	}
	if err != nil {
		return err
	}

	if err = os.Chdir(cwd); err != nil {
		return err
	}

	err = m.Commands.Post.Invoke()
	if err != nil {
		return err
	}
	return nil
}

//
// copyAndCalcHash() is based on io.copyBuffer()
//   https://golang.org/src/io/io.go
// Copyright 2009 The Go Authors. All rights reserved.
//
func (m *Manifest) copyAndCalcHash(dst io.Writer, src io.Reader) (written int64, sum string, err error) {
	h, err := m.newHash()
	if err != nil {
		return int64(0), "", err
	}
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			h.Write(buf[0:nr])
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	s := fmt.Sprintf("%x", h.Sum(nil))
	return written, s, err
}

func ParseManifest(data []byte) (*Manifest, error) {
	m := &Manifest{}
	if err := yaml.Unmarshal(data, m); err != nil {
		return nil, err
	}
	if m.Src == "" {
		return nil, fmt.Errorf("Src is required")
	}
	if m.Dest == "" {
		return nil, fmt.Errorf("Dest is required")
	}
	if m.DestMode == nil {
		mode := DefaultDestMode
		m.DestMode = &mode
	}
	return m, nil
}
