package stretcher_test

import (
	"bytes"
	"testing"

	"github.com/fujiwara/stretcher"
)

var ConsulInput1 = `[]`
var ConsulInput2 = `
[
  {
    "LTime": 2,
    "Version": 1,
    "TagFilter": "",
    "ServiceFilter": "",
    "NodeFilter": "",
    "Payload": "czM6Ly9leGFtcGxlLmNvbS9wYXRoL3RvL3Rhci5neg==",
    "Name": "deploy",
    "ID": "1d6731a8-833c-1aff-94e5-aa5e5a77da9f"
  },
  {
    "LTime": 3,
    "Version": 1,
    "TagFilter": "",
    "ServiceFilter": "",
    "NodeFilter": "",
    "Payload": "czM6Ly9leGFtcGxlLmNvbS9wYXRoL3RvL2FwcC50YXIuZ3o=",
    "Name": "deploy",
    "ID": "b5ef1588-1bcd-d93f-5d9c-67cb6e8c4587"
  }
]
`

func TestParseConsulEvents1(t *testing.T) {
	in := bytes.NewReader([]byte(ConsulInput1))
	ev, err := stretcher.ParseConsulEvents(in)
	if err == nil {
		t.Error(err)
	}
	if ev != nil {
		t.Error("Input1 must be empty!")
	}
}

func TestParseConsulEvents2(t *testing.T) {
	in := bytes.NewReader([]byte(ConsulInput2))
	ev, err := stretcher.ParseConsulEvents(in)
	if err != nil {
		t.Error(err)
	}
	if ev.ID != "b5ef1588-1bcd-d93f-5d9c-67cb6e8c4587" {
		t.Error("invalid ID")
	}
	if ev.Name != "deploy" {
		t.Error("invalid Name")
	}
	if ev.PayloadString() != "s3://example.com/path/to/app.tar.gz" {
		t.Error("invalid PayloadString()", ev.PayloadString())
	}
}
