package stretcher_test

import (
	"github.com/fujiwara/stretcher"
	"testing"
)

func TestLoadAWSConfigFile(t *testing.T) {
	stretcher.LoadAWSConfigFile("test/aws.config")
	if stretcher.AWSAuth.AccessKey != "DZECFQXRXAPORD3VDEKG" {
		t.Errorf("access_key mismatch")
	}
	if stretcher.AWSAuth.SecretKey != "NEZ5eNlmtE6ZjEQlj/uJpsog2ndjPX+uej1CHMYH" {
		t.Errorf("secret_key mismatch")

	}
	if stretcher.AWSRegion.Name != "ap-northeast-1" {
		t.Errorf("region mismatch")
	}
}
