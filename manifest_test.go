package stretcher_test

import (
	"fmt"
	"github.com/fujiwara/stretcher"
	"io/ioutil"
	"os"
	"testing"
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
	_testDest, _ := ioutil.TempFile(os.TempDir(), "stretcher_dest")
	testDest := _testDest.Name()
	os.Remove(testDest)
	os.Mkdir(testDest, 0755)
	defer os.RemoveAll(testDest)
	defer os.Remove("test/tmp/pre.touch")
	defer os.Remove("test/tmp/post.touch")

	// touch pid file (must not be deleted)
	ioutil.WriteFile(
		testDest+"/test.pid",
		[]byte(fmt.Sprintf("%d", os.Getpid())),
		0644,
	)

	cwd, _ := os.Getwd()
	yml := `
src: file://` + cwd + `/test/test.tar
checksum: 7b57db167410e46720b1d636ee6cb6c147efac3a
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
	_, err = os.Open(testDest + "/test.pid")
	if err == nil {
		t.Error("test.pid must be removed")
	}
}

func TestDeployManifestExclude(t *testing.T) {
	_testDest, _ := ioutil.TempFile(os.TempDir(), "stretcher_dest")
	testDest := _testDest.Name()
	os.Remove(testDest)
	os.Mkdir(testDest, 0755)
	defer os.RemoveAll(testDest)

	// touch pid file (must not be deleted)
	ioutil.WriteFile(
		testDest+"/test.pid",
		[]byte(fmt.Sprintf("%d", os.Getpid())),
		0644,
	)

	cwd, _ := os.Getwd()
	yml := `
src: file://` + cwd + `/test/test.tar
checksum: 7b57db167410e46720b1d636ee6cb6c147efac3a
dest: ` + testDest + `
excludes:
  - "*.pid"
  - baz
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	err = m.Deploy()
	if err != nil {
		t.Error(err)
	}
	if _, err := os.Open(testDest + "/foo/baz"); err == nil {
		t.Error("/foo/baz must be excluded")
	}
	if _, err := os.Open(testDest + "/bar"); err != nil {
		t.Error(err)
	}
	if _, err := os.Open(testDest + "/test.pid"); err != nil {
		t.Error(err)
	}
}

func TestDeployManifestExcludeFrom(t *testing.T) {
	_testDest, _ := ioutil.TempFile(os.TempDir(), "stretcher_dest")
	testDest := _testDest.Name()
	os.Remove(testDest)
	os.Mkdir(testDest, 0755)
	defer os.RemoveAll(testDest)

	// touch pid file (must not be deleted)
	ioutil.WriteFile(
		testDest+"/test.pid",
		[]byte(fmt.Sprintf("%d", os.Getpid())),
		0644,
	)

	cwd, _ := os.Getwd()
	yml := `
src: file://` + cwd + `/test/test.tar
checksum: 7b57db167410e46720b1d636ee6cb6c147efac3a
dest: ` + testDest + `
exclude_from: exclude.txt
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	err = m.Deploy()
	if err != nil {
		t.Error(err)
	}
	if _, err := os.Open(testDest + "/foo/baz"); err == nil {
		t.Error("/foo/baz must be excluded")
	}
	if _, err := os.Open(testDest + "/bar"); err != nil {
		t.Error(err)
	}
	if _, err := os.Open(testDest + "/test.pid"); err != nil {
		t.Error(err)
	}
}
