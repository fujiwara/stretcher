package stretcher

import (
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"io/ioutil"
	"os"
	"strings"
)

func LoadAWSConfigFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Cannot open %s %s", file, err)
	}
	defer f.Close()
	body, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("Cannot read %s %s", file, err)
	}

	config := map[string]string{}
	for _, line := range strings.Split(string(body), "\n") {
		if strings.Index(line, "=") != -1 {
			pair := strings.Split(line, "=")
			key := strings.Trim(pair[0], " ")
			val := strings.Trim(pair[1], " ")
			config[key] = val
		}
	}
	AWSAuth = aws.Auth{
		AccessKey: config["aws_access_key_id"],
		SecretKey: config["aws_secret_access_key"],
	}
	AWSRegion = aws.GetRegion(config["region"])
	Logger.Printf("aws_access_key_id=%s", AWSAuth.AccessKey)
	Logger.Printf("region=%s", AWSRegion.Name)
	return nil
}
