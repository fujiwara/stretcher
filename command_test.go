package stretcher_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fujiwara/stretcher"
)

func TestCommandLines(t *testing.T) {
	stretcher.Init(time.Duration(0))
	now := time.Now()
	ymd := now.Format("20060102")
	hm := now.Format("1504")
	cmdlines := stretcher.CommandLines{
		stretcher.CommandLine("date +%Y%m%d"),
		stretcher.CommandLine("date +%H%M"),
	}
	cmdlines.Invoke()
	output := stretcher.LogBuffer.String()
	if strings.Index(output, ymd) == -1 {
		t.Error("invalid output", output)
	}
	if strings.Index(output, hm) == -1 {
		t.Error("invalid output", output)
	}
}

func TestCommandLinesPipe(t *testing.T) {
	stretcher.Init(time.Duration(0))
	var buf bytes.Buffer
	for i := 0; i < 10; i++ {
		buf.WriteString(fmt.Sprintf("foo%d\n", i))
	}
	toWrite := buf.Bytes()

	cmdlines := stretcher.CommandLines{
		stretcher.CommandLine("cat > test/tmp/cmdoutput"),
		stretcher.CommandLine("cat"),
		stretcher.CommandLine("echo ok"),
	}
	defer os.Remove("test/tmp/cmdoutput")

	err := cmdlines.InvokePipe(&buf)
	if err != nil {
		t.Error(err)
	}

	wrote, err := ioutil.ReadFile("test/tmp/cmdoutput")
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(toWrite, wrote) != 0 {
		t.Error("unexpected wrote data", wrote)
	}
}

func TestCommandLinesFail(t *testing.T) {
	stretcher.Init(time.Duration(0))
	cmdlines := stretcher.CommandLines{
		stretcher.CommandLine("echo 'FOO'; echo 'BAR' 2>&1; false"),
	}
	err := cmdlines.Invoke()
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
	stretcher.Init(time.Duration(0))
	var buf bytes.Buffer
	for i := 0; i < 1025; i++ {
		buf.WriteString("0123456789012345678901234567890123456789012345678901234567890123") // 64 bytes
	}
	cmdlines := stretcher.CommandLines{
		stretcher.CommandLine("echo ok"),
	}

	err := cmdlines.InvokePipe(&buf)
	if err != nil {
		t.Error(err)
	}
	if strings.Index(stretcher.LogBuffer.String(), "broken pipe") != -1 {
		t.Error("broken pipe was occuered")
	}
}
