package io

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	iam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	ini "gopkg.in/ini.v1"
)

func getHomeDirectory() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func readAWSIniFile(filename string) *ini.File {
	homeDirectory := getHomeDirectory()
	location := filepath.Join(homeDirectory, ".aws", filename)
	_, err := os.Stat(location)
	if err != nil {
		log.Fatal("No file found at", location)
	}
	cfg, err := ini.Load(location)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func writeAWSIniFile(filename string, cfg *ini.File) string {
	homeDirectory := getHomeDirectory()
	location := filepath.Join(homeDirectory, ".aws", filename)
	cfg.SaveTo(location)

	return location
}

func updateCredentials(section string, parameters map[string]string) string {
	cfg := readAWSIniFile("credentials")
	for k, v := range parameters {
		cfg.Section(section).Key(k).SetValue(v)
	}
	location := writeAWSIniFile("credentials", cfg)

	return location
}

func getMfaSerialNumber() string {
	cfg := readAWSIniFile("config")

	return cfg.Section("default").Key("mfa_serial").String()
}

func getCurrentIamKey() string {
	cfg := readAWSIniFile("credentials")
	key := cfg.Section("default_original").Key("aws_access_key_id").String()
	if key == "" {
		log.Fatal("failed to fetch aws_access_key_id from default_original section in ~/.aws/credentials")
	}
	return key
}

func deleteCurrentIamKey() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           "default",
	}))
	client := iam.New(sess)
	output, err := client.ListAccessKeys(&iam.ListAccessKeysInput{
		UserName: nil,
	})
	if err != nil {
		log.Fatal(err)
	}

	usedKey := getCurrentIamKey()
	for index := 0; index < len(output.AccessKeyMetadata); index++ {
		candidate := *output.AccessKeyMetadata[index].AccessKeyId
		if candidate != usedKey {
			continue
		}
		_, err := client.DeleteAccessKey(&iam.DeleteAccessKeyInput{
			AccessKeyId: &candidate,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getTokenCode() string {
	var tokenCode string
	fmt.Print("Please enter mfa code: ")
	fmt.Scanln(&tokenCode)

	return tokenCode
}

func getSessionKeys(profile string) map[string]string {
	tokenCode := getTokenCode()
	serialNumber := getMfaSerialNumber()
	sessionLifespan := int64(60 * 60 * 36)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
	}))
	client := sts.New(sess)
	output, err := client.GetSessionToken(&sts.GetSessionTokenInput{
		SerialNumber:    &serialNumber,
		TokenCode:       &tokenCode,
		DurationSeconds: &sessionLifespan,
	})
	cfg := readAWSIniFile("credentials")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "InvalidClientTokenId":
				accessKey := cfg.Section(profile).Key("aws_access_key_id").String()
				log.Fatal("check your security credentials at https://console.aws.amazon.com/iam/home to ensure that access key ", accessKey, " is active")
			default:
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}

	params := make(map[string]string)
	params["aws_access_key_id"] = *output.Credentials.AccessKeyId
	params["aws_secret_access_key"] = *output.Credentials.SecretAccessKey
	params["aws_session_token"] = *output.Credentials.SessionToken
	return params
}

func getNewIamKey(profile string) map[string]string {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
	}))
	client := iam.New(sess)
	newKeyOutput, err := client.CreateAccessKey(&iam.CreateAccessKeyInput{})
	if err != nil {
		log.Fatal(err)
	}
	params := make(map[string]string)
	params["aws_access_key_id"] = *newKeyOutput.AccessKey.AccessKeyId
	params["aws_secret_access_key"] = *newKeyOutput.AccessKey.SecretAccessKey

	return params
}

// UpdateSessionKeys uses the default_original profile to get new session keys
// for the default profile
func UpdateSessionKeys() string {
	params := getSessionKeys("default_original")
	location := updateCredentials("default", params)

	return location
}

// RotateAccessKeys uses the default profile to get new access keys for the
// default_original profile. In the process, it deletes the current default_original access key
func RotateAccessKeys() string {
	deleteCurrentIamKey()
	params := getNewIamKey("default")
	location := updateCredentials("default_original", params)

	return location
}
