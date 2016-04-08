package stretcher

import (
	"log"
	"os/exec"
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
