package stretcher_test

import (
	"bytes"
	"fmt"
	"github.com/fujiwara/stretcher"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCommandLines(t *testing.T) {
	stretcher.Init()
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
	stretcher.Init()
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
