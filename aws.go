package stretcher

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/AdRoll/goamz/aws"
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
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		awsRegion = aws.GetRegion(region)
	}

	// 1. from ENV
	if auth, _ := aws.EnvAuth(); isValidAuth(auth) {
		awsAuth = auth
	}
	if isValidAuth(awsAuth) && isValidRegion(awsRegion) {
		return awsAuth, awsRegion, nil
	}

	// 2. from File (~/.aws/config, ~/.aws/credentials)
	if f := os.Getenv("AWS_CONFIG_FILE"); f != "" {
		dir, _ := filepath.Split(f)
		profile := AWSDefaultProfileName
		if profileName != AWSDefaultProfileName {
			profile = "profile " + profileName
		}
		auth, region, _ := loadAWSConfigFile(f, profile)
		if isValidAuth(auth) {
			awsAuth = auth
		}
		if isValidRegion(region) {
			awsRegion = region
		}

		cred := filepath.Join(dir, "credentials")
		auth, region, _ = loadAWSConfigFile(cred, profileName)
		if isValidAuth(auth) {
			awsAuth = auth
		}
		if isValidRegion(region) {
			awsRegion = region
		}
	}
	if isValidAuth(awsAuth) && isValidRegion(awsRegion) {
		return awsAuth, awsRegion, nil
	}

	// 3. from IAM Role
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
