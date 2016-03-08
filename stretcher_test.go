package stretcher_test

import (
	"testing"
	"time"

	"github.com/fujiwara/stretcher"
)

func TestRandomTime(t *testing.T) {
	var sum time.Duration
	for i := 0; i < 10000; i++ {
		s := stretcher.RandomTime(10)
		e := time.Duration(10) * time.Second
		if s > e {
			t.Errorf("RandomTime too long. expected %s got %s", e, s)
		}
		sum += s
	}
	avg := sum / 10000
	if avg < time.Duration(4)*time.Second || time.Duration(6)*time.Second < avg {
		t.Errorf("average of random time(0 ~ 10s) is out of range (%s)", avg)
	}
}

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
