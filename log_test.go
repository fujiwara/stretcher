package stretcher_test

import (
	"bytes"
	"github.com/fujiwara/stretcher"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := stretcher.NewLogger(&buf)
	logger.Println("foo", "bar")
	logger.Printf("aaa=%d", 123)
	str := buf.String()
	if strings.Index(str, "foo bar") == -1 {
		t.Error("logger output mismatch")
	}
	if strings.Index(str, "aaa=123") == -1 {
		t.Error("logger output mismatch")
	}
}
