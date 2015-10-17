package stretcher

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/AdRoll/goamz/aws"
	homedir "github.com/mitchellh/go-homedir"
	ini "github.com/vaughan0/go-ini"
)

const (
	AWSDefaultRegionName  = "us-east-1"
	AWSDefaultProfileName = "default"
)

func isValidAuth(auth aws.Auth) bool {
	return auth.AccessKey != "" && auth.SecretKey != ""
}

func isValidRegion(region aws.Region) bool {
	return region.Name != ""
}

func LoadAWSCredentials(profileName string) (aws.Auth, aws.Region, error) {
	if profileName == "" {
		if p := os.Getenv("AWS_DEFAULT_PROFILE"); p != "" {
			profileName = p
		} else {
			profileName = AWSDefaultProfileName
		}
	}

	var awsAuth aws.Auth
	var awsRegion aws.Region

	// load from File (~/.aws/config, ~/.aws/credentials)
	configFile := os.Getenv("AWS_CONFIG_FILE")
	if configFile == "" {
		if dir, err := homedir.Dir(); err == nil {
			configFile = filepath.Join(dir, ".aws", "config")
		}
	}

	dir, _ := filepath.Split(configFile)
	_profile := AWSDefaultProfileName
	if profileName != AWSDefaultProfileName {
		_profile = "profile " + profileName
	}
	auth, region, _ := loadAWSConfigFile(configFile, _profile)
	if isValidAuth(auth) {
		awsAuth = auth
	}
	if isValidRegion(region) {
		awsRegion = region
	}

	credFile := filepath.Join(dir, "credentials")
	auth, region, _ = loadAWSConfigFile(credFile, profileName)
	if isValidAuth(auth) {
		awsAuth = auth
	}
	if isValidRegion(region) {
		awsRegion = region
	}

	// Override by environment valiable
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		awsRegion = aws.GetRegion(region)
	}
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		if auth, _ := aws.EnvAuth(); isValidAuth(auth) {
			awsAuth = auth
		}
	}
	if isValidAuth(awsAuth) && isValidRegion(awsRegion) {
		return awsAuth, awsRegion, nil
	}

	// Otherwise, use IAM Role
	cred, err := aws.GetInstanceCredentials()
	if err == nil {
		exptdate, err := time.Parse("2006-01-02T15:04:05Z", cred.Expiration)
		if err == nil {
			auth := aws.NewAuth(cred.AccessKeyId, cred.SecretAccessKey, cred.Token, exptdate)
			awsAuth = *auth
		}
	}
	if isValidAuth(awsAuth) && isValidRegion(awsRegion) {
		return awsAuth, awsRegion, nil
	}

	return awsAuth, awsRegion, errors.New("cannot detect valid credentials or region")
}

func loadAWSConfigFile(fileName string, profileName string) (aws.Auth, aws.Region, error) {
	var auth aws.Auth
	var region aws.Region

	conf, err := ini.LoadFile(fileName)
	if err != nil {
		return auth, region, err
	}
	log.Printf("Loading file %s [%s]", fileName, profileName)

	for key, value := range conf[profileName] {
		switch key {
		case "aws_access_key_id":
			auth.AccessKey = value
		case "aws_secret_access_key":
			auth.SecretKey = value
		case "region":
			region = aws.GetRegion(value)
		}
	}
	return auth, region, nil
}
