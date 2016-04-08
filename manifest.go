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
	"github.com/fujiwara/shapeio"
	"gopkg.in/yaml.v2"
)

var DefaultDestMode = os.FileMode(0755)

type Manifest struct {
	Src          string       `yaml:"src"`
	CheckSum     string       `yaml:"checksum"`
	Dest         string       `yaml:"dest"`
	DestMode     *os.FileMode `yaml:"dest_mode"`
	Commands     Commands     `yaml:"commands"`
	Excludes     []string     `yaml:"excludes"`
	ExcludeFrom  string       `yaml:"exclude_from"`
	SyncStrategy string       `yaml:"sync_strategy"`
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

func (m *Manifest) Deploy(conf Config) error {
	tmp, err := ioutil.TempFile(os.TempDir(), "stretcher")
	if err != nil {
		return err
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	if conf.Timeout != 0 {
		log.Printf("Set timeout %s", conf.Timeout)
		timer := time.NewTimer(conf.Timeout)
		done := make(chan error)
		go func() {
			done <- m.fetchSrc(conf, tmp)
		}()
		select {
		case <-timer.C:
			return fmt.Errorf("timeout %s reached while fetching src %s", conf.Timeout, m.Src)
		case err := <-done:
			if err != nil {
				return err
			}
		}
	} else {
		err := m.fetchSrc(conf, tmp)
		if err != nil {
			return err
		}
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

	strategy, err := NewSyncStrategy(m)
	if err != nil {
		return err
	}

	err = strategy.Sync(from, to)
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

func (m *Manifest) fetchSrc(conf Config, tmp *os.File) error {
	begin := time.Now()
	src, err := getURL(m.Src)
	if err != nil {
		return fmt.Errorf("Get src failed: %s", err)
	}
	defer src.Close()

	lsrc := shapeio.NewReader(src)
	if conf.MaxBandWidth != 0 {
		log.Printf("Set max bandwidth %s/sec", humanize.Bytes(uint64(conf.MaxBandWidth)))
		lsrc.SetRateLimit(float64(conf.MaxBandWidth))
	}

	written, sum, err := m.copyAndCalcHash(tmp, lsrc)
	if err != nil {
		return err
	}
	elapsed := time.Since(begin)
	log.Printf("Wrote %s bytes to %s (in %s, %s/s)",
		humanize.Comma(written),
		tmp.Name(),
		elapsed,
		humanize.Bytes(uint64(float64(written)/elapsed.Seconds())),
	)
	if len(m.CheckSum) > 0 && sum != strings.ToLower(m.CheckSum) {
		return fmt.Errorf("Checksum mismatch. expected:%s got:%s", m.CheckSum, sum)
	} else {
		log.Printf("Checksum ok: %s", sum)
	}
	return nil
}

func (m *Manifest) copyAndCalcHash(dst io.Writer, src io.Reader) (int64, string, error) {
	h, err := m.newHash()
	if err != nil {
		return 0, "", err
	}
	w := io.MultiWriter(h, dst)

	written, err := io.Copy(w, src)
	if err != nil {
		return written, "", err
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
