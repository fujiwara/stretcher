package stretcher

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/crowdmob/goamz/aws"
)

const (
	AWSDefaultRegionName  = "us-east-1"
	AWSDefaultProfileName = "default"
)

func LoadAWSConfigFile(fileName string, profileName string) error {
	if profileName == "" {
		profileName = "default"
	}

	f, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("Cannot open %s %s", fileName, err)
	}
	defer f.Close()
	body, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("Cannot read %s %s", fileName, err)
	}

	profiles := map[string]map[string]string{}
	currentProfile := ""
	for _, line := range strings.Split(string(body), "\n") {
		if strings.Index(line, "[profile ") == 0 {
			p := strings.Split(line, " ")
			if len(p) < 2 {
				continue
			}
			currentProfile = strings.Replace(p[1], "]", "", 1)
		} else if strings.Index(line, "[default]") == 0 {
			currentProfile = "default"
		} else if strings.Index(line, "=") != -1 {
			pair := strings.Split(line, "=")
			key := strings.Trim(pair[0], " ")
			val := strings.Trim(pair[1], " ")
			if _, ok := profiles[currentProfile]; !ok {
				profiles[currentProfile] = map[string]string{}
			}
			profiles[currentProfile][key] = val
		}
	}
	profile, ok := profiles[profileName]
	if !ok {
		return fmt.Errorf("profile [%s] not found in %s", profileName, fileName)
	}
	AWSAuth = aws.Auth{
		AccessKey: profile["aws_access_key_id"],
		SecretKey: profile["aws_secret_access_key"],
	}
	if profile["region"] == "" {
		profile["region"] = AWSDefaultRegionName
	}
	AWSRegion = aws.GetRegion(profile["region"])
	log.Printf("aws_access_key_id=%s", AWSAuth.AccessKey)
	log.Printf("region=%s", AWSRegion.Name)
	return nil
}
