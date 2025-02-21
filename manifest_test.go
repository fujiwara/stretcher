package stretcher_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fujiwara/stretcher"
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

func TestParseManifestWithEnv(t *testing.T) {
	t.Setenv("SRC", "s3://example.com/path/to/archive.tar.gz")
	yml := `
src: '{{ must_env "SRC" }}'
checksum: e0840daaa97cd2cf2175f9e5d133ffb3324a2b93
dest: '{{ env "DEST" "/home/stretcher/app" }}'
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
		t.Errorf("invalid src: %s", m.Src)
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

func TestParseManifestSyncStrategy(t *testing.T) {
	yml := `
src: s3://example.com/path/to/archive.tar.gz
checksum: e0840daaa97cd2cf2175f9e5d133ffb3324a2b93
dest: /home/stretcher/app
sync_strategy: mv
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
	if m.SyncStrategy != "mv" {
		t.Error("invalid sync_strategy", m.SyncStrategy)
	}
}

func TestDeployManifest(t *testing.T) {
	ctx := context.Background()
	_testDest, _ := os.CreateTemp(os.TempDir(), "stretcher_dest")
	testDest := _testDest.Name()
	os.Remove(testDest)
	os.Mkdir(testDest, 0755)
	defer os.RemoveAll(testDest)
	defer os.Remove("testdata/tmp/pre.touch")
	defer os.Remove("testdata/tmp/post.touch")

	// touch pid file (must not be deleted)
	os.WriteFile(
		testDest+"/test.pid",
		[]byte(fmt.Sprintf("%d", os.Getpid())),
		0644,
	)

	cwd, _ := os.Getwd()
	yml := `
src: file://` + cwd + `/testdata/test.tar
checksum: 7b57db167410e46720b1d636ee6cb6c147efac3a
dest: ` + testDest + `
commands:
  pre:
    - pwd
    - echo "pre" > testdata/tmp/pre.touch
  post:
    - pwd
    - echo "post" > testdata/tmp/post.touch
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	err = m.Deploy(ctx, stretcher.Config{})
	if err != nil {
		t.Error(err)
	}
	if _, err := os.Open(testDest + "/foo/baz"); err != nil {
		t.Error(err)
	}
	if _, err := os.Open(testDest + "/bar"); err != nil {
		t.Error(err)
	}
	if _, err := os.Open("testdata/tmp/pre.touch"); err != nil {
		t.Error(err)
	}
	if _, err := os.Open("testdata/tmp/post.touch"); err != nil {
		t.Error(err)
	}
	_, err = os.Open(testDest + "/test.pid")
	if err == nil {
		t.Error("test.pid must be removed")
	}
}

func TestDeployManifestSyncStrategyMv(t *testing.T) {
	ctx := context.Background()
	_testDest, _ := os.CreateTemp(os.TempDir(), "stretcher_dest")
	testDest := _testDest.Name()
	//	os.Remove(testDest)
	defer os.RemoveAll(testDest)
	defer os.Remove("testdata/tmp/pre.touch")
	defer os.Remove("testdata/tmp/post.touch")

	// touch pid file (must not be deleted)
	os.WriteFile(
		testDest+"/test.pid",
		[]byte(fmt.Sprintf("%d", os.Getpid())),
		0644,
	)

	cwd, _ := os.Getwd()
	yml := `
src: file://` + cwd + `/testdata/test.tar
checksum: 7b57db167410e46720b1d636ee6cb6c147efac3a
dest: ` + testDest + `
sync_strategy: mv
dest_mode: 0751
commands:
  pre:
    - rm -fr ` + testDest + `
    - pwd
    - echo "pre" > testdata/tmp/pre.touch
  post:
    - pwd
    - echo "post" > testdata/tmp/post.touch
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	err = m.Deploy(ctx, stretcher.Config{})
	if err != nil {
		t.Error(err)
	}

	stat, err := os.Stat(testDest)
	if err != nil {
		t.Error(err)
	}
	if stat.Mode().Perm() != 0751 {
		t.Errorf("dest mode %s expected 0751", stat.Mode().Perm())
	}
	if _, err := os.Open(testDest + "/foo/baz"); err != nil {
		t.Error(err)
	}
	if _, err := os.Open(testDest + "/bar"); err != nil {
		t.Error(err)
	}
	if _, err := os.Open("testdata/tmp/pre.touch"); err != nil {
		t.Error(err)
	}
	if _, err := os.Open("testdata/tmp/post.touch"); err != nil {
		t.Error(err)
	}
	_, err = os.Open(testDest + "/test.pid")
	if err == nil {
		t.Error("test.pid must be removed")
	}
}

func TestDeployManifestInvalidSyncStrategy(t *testing.T) {
	ctx := context.Background()
	_testDest, _ := os.CreateTemp(os.TempDir(), "stretcher_dest")
	testDest := _testDest.Name()
	defer os.RemoveAll(testDest)
	cwd, _ := os.Getwd()
	yml := `
src: file://` + cwd + `/testdata/test.tar
checksum: 7b57db167410e46720b1d636ee6cb6c147efac3a
dest: ` + testDest + `
sync_strategy: dummy
commands:
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	err = m.Deploy(ctx, stretcher.Config{})
	if err == nil || strings.Index(err.Error(), "invalid strategy") == -1 {
		t.Error("error must be occured: invalid sync_strategy", err)
	}
}

func TestDeployManifestExclude(t *testing.T) {
	ctx := context.Background()
	_testDest, _ := os.CreateTemp(os.TempDir(), "stretcher_dest")
	testDest := _testDest.Name()
	os.Remove(testDest)
	os.Mkdir(testDest, 0755)
	defer os.RemoveAll(testDest)

	// touch pid file (must not be deleted)
	os.WriteFile(
		testDest+"/test.pid",
		[]byte(fmt.Sprintf("%d", os.Getpid())),
		0644,
	)

	cwd, _ := os.Getwd()
	yml := `
src: file://` + cwd + `/testdata/test.tar
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
	err = m.Deploy(ctx, stretcher.Config{})
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
	ctx := context.Background()
	_testDest, _ := os.CreateTemp(os.TempDir(), "stretcher_dest")
	testDest := _testDest.Name()
	os.Remove(testDest)
	os.Mkdir(testDest, 0755)
	defer os.RemoveAll(testDest)

	// touch pid file (must not be deleted)
	os.WriteFile(
		testDest+"/test.pid",
		[]byte(fmt.Sprintf("%d", os.Getpid())),
		0644,
	)

	cwd, _ := os.Getwd()
	yml := `
src: file://` + cwd + `/testdata/test.tar
checksum: 7b57db167410e46720b1d636ee6cb6c147efac3a
dest: ` + testDest + `
exclude_from: exclude.txt
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	err = m.Deploy(ctx, stretcher.Config{})
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

func TestDeployManifestDestMode(t *testing.T) {
	ctx := context.Background()
	_testDest, _ := os.CreateTemp(os.TempDir(), "stretcher_dest")
	testDest := _testDest.Name()
	os.Remove(testDest)
	os.Mkdir(testDest, 0755)
	defer os.RemoveAll(testDest)

	cwd, _ := os.Getwd()
	yml := `
src: file://` + cwd + `/testdata/test_no_top_dir.tar
checksum: da5ec3a7dca4b0492a0ba0104f7cc7ad2ae2eafc
dest: ` + testDest + `
dest_mode: 0711
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	err = m.Deploy(ctx, stretcher.Config{})
	if err != nil {
		t.Error(err)
	}
	stat, err := os.Stat(testDest)
	if err != nil {
		t.Error(err)
	}
	if stat.Mode().Perm() != 0711 {
		t.Errorf("dest mode %s expected 0711", stat.Mode().Perm())
	}
}

func TestParseManifestWithColonInvalid(t *testing.T) {
	yml := `
src: s3://example.com/path/to/archive.tar.gz
checksum: e0840daaa97cd2cf2175f9e5d133ffb3324a2b93
dest: /home/stretcher/app
commands:
  success:
    - some-commend-with-argument-includes-colon ":foo: bar"
`
	_, err := stretcher.ParseManifest([]byte(yml))
	if err == nil {
		t.Error("must be parse error")
	}
}

func TestParseManifestWithColonValid(t *testing.T) {
	yml := `
src: s3://example.com/path/to/archive.tar.gz
checksum: e0840daaa97cd2cf2175f9e5d133ffb3324a2b93
dest: /home/stretcher/app
commands:
  success:
    - 'some-commend-with-argument-includes-colon ":foo: bar"'
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
		return
	}
	if len(m.Commands.Success) != 1 {
		t.Errorf("invalid commands.success")
	}
	if m.Commands.Success[0] != `some-commend-with-argument-includes-colon ":foo: bar"` {
		t.Errorf("invalid commands.success[0]: %s", m.Commands.Success[0])
	}
}

func TestDeployManifestRetry(t *testing.T) {
	ctx := context.Background()
	cwd, _ := os.Getwd()
	yml := `
src: file://` + cwd + `/testdata/test_not_exist_filepath
checksum: 7b57db167410e46720b1d636ee6cb6c147efac3a
dest: ` + cwd + `/testdata/dest
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	err = m.Deploy(ctx, stretcher.Config{
		Retry:     3,
		RetryWait: 3 * time.Second,
	})
	if err == nil || !strings.Contains(err.Error(), "src failed:") {
		t.Errorf("expect retry got %s", err)
	}
}

func TestDeployManifestTimeout(t *testing.T) {
	ctx := context.Background()
	_testDest, _ := os.CreateTemp(os.TempDir(), "stretcher_dest")
	testDest := _testDest.Name()
	os.Remove(testDest)
	os.Mkdir(testDest, 0755)
	defer os.RemoveAll(testDest)
	cwd, _ := os.Getwd()
	yml := `
src: file://` + cwd + `/testdata/test.tar
checksum: 7b57db167410e46720b1d636ee6cb6c147efac3a
dest: ` + testDest + `
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	stat, err := os.Stat(cwd + "/testdata/test.tar")
	if err != nil {
		t.Error(err)
	}
	// download will be finished in about 5 seconds
	bw := uint64(stat.Size()) / 5
	err = m.Deploy(ctx, stretcher.Config{
		MaxBandWidth: bw,
		Timeout:      time.Duration(2 * time.Second),
	})
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expect timeout got %s", err)
	}
}

func TestDeployManifestMaxBandwidth(t *testing.T) {
	ctx := context.Background()
	_testDest, _ := os.CreateTemp(os.TempDir(), "stretcher_dest")
	testDest := _testDest.Name()
	os.Remove(testDest)
	os.Mkdir(testDest, 0755)
	defer os.RemoveAll(testDest)
	cwd, _ := os.Getwd()
	yml := `
src: file://` + cwd + `/testdata/test.tar
checksum: 7b57db167410e46720b1d636ee6cb6c147efac3a
dest: ` + testDest + `
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	stat, err := os.Stat(cwd + "/testdata/test.tar")
	if err != nil {
		t.Error(err)
	}
	// expect to download in about 2 second
	bw := uint64(stat.Size()) / 2
	start := time.Now()
	err = m.Deploy(ctx, stretcher.Config{
		MaxBandWidth: bw,
	})
	if err != nil {
		t.Error(err)
	}
	elapsed := time.Since(start)
	if elapsed.Seconds() < 1 || 4 < elapsed.Seconds() {
		t.Errorf("elapsed expect about 2 sec. but %s", elapsed)
	}
}

func TestDeployCommandsSuccess(t *testing.T) {
	ctx := context.Background()
	yml := `
commands:
  pre:
    - 'echo "pre 1"'
    - 'echo "pre 2"'
  post:
    - 'echo "post 1"'
    - 'echo "post 2"'
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	err = m.Deploy(ctx, stretcher.Config{})
	if err != nil {
		t.Error(err)
	}
}

func TestDeployCommandsFail(t *testing.T) {
	ctx := context.Background()
	yml := `
commands:
  pre:
    - 'echo "pre 1"'
    - exit 1
  post:
    - 'echo "post 1"'
    - 'echo "post 2"'
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	err = m.Deploy(ctx, stretcher.Config{})
	if err == nil {
		t.Error(err)
	}
}

func TestDeployContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	yml := `
commands:
  pre:
    - 'echo "pre 1"'
`
	m, err := stretcher.ParseManifest([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	cancel() // cancel context before deploy
	err = m.Deploy(ctx, stretcher.Config{})
	if err == nil {
		t.Error("context canceled error must be raised")
	}
}
