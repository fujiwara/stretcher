package stretcher_test

import (
	"testing"

	"github.com/fujiwara/stretcher"
)

func TestLoadAWSConfigFile(t *testing.T) {
	stretcher.LoadAWSConfigFile("test/aws.config", "default")
	if stretcher.AWSAuth.AccessKey != "DZECFQXRXAPORD3VDEKG" {
		t.Errorf("access_key mismatch")
	}
	if stretcher.AWSAuth.SecretKey != "NEZ5eNlmtE6ZjEQlj/uJpsog2ndjPX+uej1CHMYH" {
		t.Errorf("secret_key mismatch")

	}
	if stretcher.AWSRegion.Name != "ap-northeast-1" {
		t.Errorf("region mismatch")
	}

	stretcher.LoadAWSConfigFile("test/aws.config", "dummy")
	if stretcher.AWSAuth.AccessKey != "AAAAAAAAAAAAAAAAAAAA" {
		t.Errorf("access_key mismatch")
	}
	if stretcher.AWSAuth.SecretKey != "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB" {
		t.Errorf("secret_key mismatch")

	}
	if stretcher.AWSRegion.Name != "us-east-1" {
		t.Errorf("region mismatch")
	}

	err := stretcher.LoadAWSConfigFile("test/aws.config", "not_found")
	if err == nil {
		t.Error("load not_found profile must be error", err)
	}
}
