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
	cmdlines.Invoke(ctx, nil)
	output := stretcher.LogBuffer.String()
	if !strings.Contains(output, ymd) {
		t.Error("invalid output", output)
	}
	if !strings.Contains(output, hm) {
		t.Error("invalid output", output)
	}
}

func TestCommandLinesWithEnv(t *testing.T) {
	ctx := context.Background()
	stretcher.LogBuffer.Reset()
	now := time.Now()
	ymd := now.Format("20060102")
	hm := now.Format("1504")
	cmdlines := stretcher.CommandLines{
		stretcher.CommandLine("env | grep STRETCHER_TEST_"),
	}
	cmdlines.Invoke(ctx, map[string]string{
		"STRETCHER_TEST_YMD": ymd,
		"STRETCHER_TEST_HM":  hm,
	})
	output := stretcher.LogBuffer.String()
	if !strings.Contains(output, "STRETCHER_TEST_YMD="+ymd) {
		t.Error("invalid output", output)
	}
	if !strings.Contains(output, "STRETCHER_TEST_HM="+hm) {
		t.Error("invalid output", output)
	}
}

func TestCommandLinesPipeWithEnv(t *testing.T) {
	ctx := context.Background()
	stretcher.LogBuffer.Reset()
	cmdlines := stretcher.CommandLines{
		stretcher.CommandLine(`perl -E '<>; say $ENV{STRETCHER_TEST}' > testdata/tmp/cmdoutput`),
	}
	in := new(bytes.Buffer)
	cmdlines.InvokePipe(ctx, in, map[string]string{
		"STRETCHER_TEST": "hello_env",
	})
	defer os.Remove("testdata/tmp/cmdoutput")
	b, err := os.ReadFile("testdata/tmp/cmdoutput")
	if err != nil {
		t.Error(err)
	}
	output := string(b)
	t.Log("OUTPUT=", output)
	if !strings.Contains(output, "hello_env") {
		t.Error("invalid output", output)
	}
}

func TestCommandLinesPipe(t *testing.T) {
	ctx := context.Background()
	stretcher.LogBuffer.Reset()
	var buf bytes.Buffer
	for i := range 10 {
		buf.WriteString(fmt.Sprintf("foo%d\n", i))
	}
	toWrite := buf.Bytes()

	cmdlines := stretcher.CommandLines{
		stretcher.CommandLine("cat > testdata/tmp/cmdoutput"),
		stretcher.CommandLine("cat"),
		stretcher.CommandLine("echo ok"),
	}
	defer os.Remove("testdata/tmp/cmdoutput")

	err := cmdlines.InvokePipe(ctx, &buf, nil)
	if err != nil {
		t.Error(err)
	}

	wrote, err := os.ReadFile("testdata/tmp/cmdoutput")
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(toWrite, wrote) {
		t.Error("unexpected wrote data", wrote)
	}
}

func TestCommandLinesFail(t *testing.T) {
	ctx := context.Background()
	stretcher.LogBuffer.Reset()
	cmdlines := stretcher.CommandLines{
		stretcher.CommandLine("echo 'FOO'; echo 'BAR' 2>&1; false"),
	}
	err := cmdlines.Invoke(ctx, nil)
	if err == nil {
		t.Error("false command must fail")
	}
	if fmt.Sprintf("%s", err) != "failed: echo 'FOO'; echo 'BAR' 2>&1; false exit status 1" {
		t.Error("invalid err message.", err)
	}
	output := stretcher.LogBuffer.String()
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

	err := cmdlines.InvokePipe(ctx, &buf, nil)
	if err != nil {
		t.Error(err)
	}
	if strings.Contains(stretcher.LogBuffer.String(), "broken pipe") {
		t.Error("broken pipe was occuered")
	}
}
