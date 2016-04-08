package stretcher

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

type SyncStrategy interface {
	Sync(from, to string) error
}

type RsyncStrategy struct {
	*Manifest
}

var RsyncDefaultOpts = []string{"-av", "--delete"}

func (s *RsyncStrategy) Sync(from, to string) error {
	m := s.Manifest

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
	out, err := exec.Command("rsync", args...).CombinedOutput()
	if len(out) > 0 {
		log.Println(string(out))
	}
	if err != nil {
		return err
	}

	return nil
}

type MvSyncStrategy struct {
}

func (s *MvSyncStrategy) Sync(from, to string) error {
	log.Printf("Rename srcdir %s to dest %s\n", from, to)
	return os.Rename(from, to)
}
