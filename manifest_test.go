package stretcher_test

import (
	"testing"
	"github.com/fujiwara/stretcher"
	"os"
	"io/ioutil"
)

func TestParseManifest(t *testing.T) {
	yml := `
src: s3://example.com/path/to/archive.tar.gz
checksum: e0840daaa97cd2cf2175f9e5d133ffb3324a2b93
dest: /home/stretcher/app
commands:
  pre:
    - echo 'staring deploy'
    - echo 'xxx'
  post:
    - echo 'deploy done'
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	if m.Src != "s3://example.com/path/to/archive.tar.gz" {
		t.Errorf("invalid src")
	}
	if m.CheckSum != "e0840daaa97cd2cf2175f9e5d133ffb3324a2b93" {
		t.Errorf("invalid checksum")
	}
	if len(m.Commands.Pre) != 2 {
		t.Errorf("invalid commands.pre")
	}
	if len(m.Commands.Post) != 1 {
		t.Errorf("invalid commands.post")
	}
}


func TestDeployManifest(t *testing.T) {
	_testDest, _ := ioutil.TempFile(os.TempDir(), "stretcher_test")
	testDest := _testDest.Name()
	os.Remove(testDest)
	os.Mkdir(testDest, 0755)
	defer os.RemoveAll(testDest)
	defer os.Remove("test/tmp/pre.touch")
	defer os.Remove("test/tmp/post.touch")

	cwd, _ := os.Getwd()
	yml := `
src: file://` + cwd + `/test/test.tar
checksum: 7e3e9491ca2d825ffa3a0f78acea3a971392bc9d
dest: ` + testDest + `
commands:
  pre:
    - pwd
    - echo "pre" > test/tmp/pre.touch
  post:
    - pwd
    - echo "post" > test/tmp/post.touch
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	err = m.Deploy()
	if err != nil {
		t.Error(err)
	}
	if _, err := os.Open(testDest + "/foo/baz"); err != nil {
		t.Error(err)
	}
	if _, err := os.Open(testDest + "/bar"); err != nil {
		t.Error(err)
	}
	if _, err := os.Open("test/tmp/pre.touch"); err != nil {
		t.Error(err)
	}
	if _, err := os.Open("test/tmp/post.touch"); err != nil {
		t.Error(err)
	}
}

