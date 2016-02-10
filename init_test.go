package stretcher_test

import (
	"testing"
	"time"

	"github.com/fujiwara/stretcher"
)

func TestInitSleep(t *testing.T) {
	sleep := time.Duration(1 * time.Second)
	start := time.Now()
	stretcher.Init(sleep)
	end := time.Now()
	diff := end.Sub(start)
	if diff < sleep {
		t.Errorf("sleeping time is not enough. expected at least %s, but returned in %s", sleep, diff)
	}
}
