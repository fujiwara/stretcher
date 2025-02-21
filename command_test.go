package stretcher_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fujiwara/stretcher"
)

func TestCommandLines(t *testing.T) {
	ctx := context.Background()
	stretcher.LogBuffer.Reset()
	now := time.Now()
	ymd := now.Format("20060102")
	hm := now.Format("1504")
	cmdlines := stretcher.CommandLines{
		stretcher.CommandLine("date +%Y%m%d"),
		stretcher.CommandLine("date +%H%M"),
	}
	cmdlines.Invoke(ctx)
	output := stretcher.LogBuffer.String()
	if strings.Index(output, ymd) == -1 {
		t.Error("invalid output", output)
	}
	if strings.Index(output, hm) == -1 {
		t.Error("invalid output", output)
	}
}

func TestCommandLinesPipe(t *testing.T) {
	ctx := context.Background()
	stretcher.LogBuffer.Reset()
	var buf bytes.Buffer
	for i := 0; i < 10; i++ {
		buf.WriteString(fmt.Sprintf("foo%d\n", i))
	}
	toWrite := buf.Bytes()

	cmdlines := stretcher.CommandLines{
		stretcher.CommandLine("cat > testdata/tmp/cmdoutput"),
		stretcher.CommandLine("cat"),
		stretcher.CommandLine("echo ok"),
	}
	defer os.Remove("testdata/tmp/cmdoutput")

	err := cmdlines.InvokePipe(ctx, &buf)
	if err != nil {
		t.Error(err)
	}

	wrote, err := os.ReadFile("testdata/tmp/cmdoutput")
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(toWrite, wrote) != 0 {
		t.Error("unexpected wrote data", wrote)
	}
}

func TestCommandLinesFail(t *testing.T) {
	ctx := context.Background()
	stretcher.LogBuffer.Reset()
	cmdlines := stretcher.CommandLines{
		stretcher.CommandLine("echo 'FOO'; echo 'BAR' 2>&1; false"),
	}
	err := cmdlines.Invoke(ctx)
	if err == nil {
		t.Error("false command must fail")
	}
	if fmt.Sprintf("%s", err) != "failed: echo 'FOO'; echo 'BAR' 2>&1; false exit status 1" {
		t.Error("invalid err message.", err)
	}
	output := string(stretcher.LogBuffer.Bytes())
	if !strings.Contains(output, "FOO\n") {
		t.Error("output does not contain FOO\\n")
	}
	if !strings.Contains(output, "BAR\n") {
		t.Error("output does not contain BAR\\n")
	}
}

func TestCommandLinesPipeIgnoreEPIPE(t *testing.T) {
	ctx := context.Background()
	stretcher.LogBuffer.Reset()
	var buf bytes.Buffer
	for i := 0; i < 1025; i++ {
		buf.WriteString("0123456789012345678901234567890123456789012345678901234567890123") // 64 bytes
	}
	cmdlines := stretcher.CommandLines{
		stretcher.CommandLine("echo ok"),
	}

	err := cmdlines.InvokePipe(ctx, &buf)
	if err != nil {
		t.Error(err)
	}
	if strings.Index(stretcher.LogBuffer.String(), "broken pipe") != -1 {
		t.Error("broken pipe was occuered")
	}
}
