package stretcher_test

import (
	"os"
	"testing"

	"github.com/fujiwara/stretcher"
)

func TestLoadAWSConfigFile(t *testing.T) {
	os.Setenv("AWS_CONFIG_FILE", "test/aws.config")

	auth, region, _ := stretcher.LoadAWSCredentials("")
	if auth.AccessKey != "DZECFQXRXAPORD3VDEKG" {
		t.Errorf("access_key mismatch")
	}
	if auth.SecretKey != "NEZ5eNlmtE6ZjEQlj/uJpsog2ndjPX+uej1CHMYH" {
		t.Errorf("secret_key mismatch")

	}
	if region.Name != "ap-northeast-1" {
		t.Errorf("region mismatch")
	}

	auth, region, _ = stretcher.LoadAWSCredentials("dummy")
	if auth.AccessKey != "AAAAAAAAAAAAAAAAAAAA" {
		t.Errorf("access_key mismatch")
	}
	if auth.SecretKey != "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB" {
		t.Errorf("secret_key mismatch")

	}
	if region.Name != "us-east-1" {
		t.Errorf("region mismatch")
	}

	os.Setenv("AWS_CONFIG_FILE", "test/aws.config")
	auth, region, err := stretcher.LoadAWSCredentials("not_found")
	if err == nil {
		t.Error("load not_found profile must be error", err)
	}

	// split config & crecdentials
	os.Setenv("AWS_CONFIG_FILE", "test/.aws/config")
	auth, region, err = stretcher.LoadAWSCredentials("foo")
	if auth.AccessKey != "foo_key" {
		t.Errorf("access_key mismatch")
	}
	if auth.SecretKey != "foo_secret" {
		t.Errorf("secret_key mismatch")
	}
	if region.Name != "us-west-2" {
		t.Errorf("region mismatch")
	}

	auth, region, err = stretcher.LoadAWSCredentials("bar")
	if auth.AccessKey != "bar_key" {
		t.Errorf("access_key mismatch")
	}
	if auth.SecretKey != "bar_secret" {
		t.Errorf("secret_key mismatch")
	}
	if region.Name != "ap-southeast-1" {
		t.Errorf("region mismatch")
	}
}
