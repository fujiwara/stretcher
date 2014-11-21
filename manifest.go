package stretcher

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"gopkg.in/yaml.v1"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var RsyncDefaultOpts = []string{"-av", "--delete"}

type Manifest struct {
	Src         string   `yaml:"src"`
	CheckSum    string   `yaml:"checksum"`
	Dest        string   `yaml:"dest"`
	Commands    Commands `yaml:"commands"`
	Excludes    []string `yaml:"excludes"`
	ExcludeFrom string   `yaml:"exclude_from"`
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
	Logger.Printf("Wrote %d bytes to %s", written, tmp.Name())
	if len(m.CheckSum) > 0 && sum != strings.ToLower(m.CheckSum) {
		return fmt.Errorf("Checksum mismatch. expected:%s got:%s", m.CheckSum, sum)
	} else {
		Logger.Printf("Checksum ok: %s", sum)
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

	cwd, _ := os.Getwd()
	if err = os.Chdir(dir); err != nil {
		return err
	}

	Logger.Println("Extract archive:", tmp.Name(), "to", dir)
	out, err := exec.Command("tar", "xf", tmp.Name()).CombinedOutput()
	if err != nil {
		Logger.Println("failed: tar xf", tmp.Name(), "failed", err)
		return err
	}
	fmt.Println(string(out))

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

	Logger.Println("rsync", args)
	out, err = exec.Command("rsync", args...).CombinedOutput()
	if err != nil {
		return err
	}
	Logger.Println(string(out))

	if err = os.Chdir(cwd); err != nil {
		return err
	}

	err = m.Commands.Post.Invoke()
	if err != nil {
		return err
	}
	return nil
}

func (m *Manifest) copyAndCalcHash(dst io.Writer, src io.Reader) (written int64, sum string, err error) {
	h, err := m.newHash()
	if err != nil {
		return int64(0), "", err
	}
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			io.WriteString(h, string(buf[0:nr]))
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
	return m, nil
}
