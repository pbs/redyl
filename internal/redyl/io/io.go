package io

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

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

func readCredentialsFile() *ini.File {
	homeDirectory := getHomeDirectory()
	credentialsPath := filepath.Join(homeDirectory, ".aws", "credentials")
	_, err := os.Stat(credentialsPath)
	if err != nil {
		log.Fatal("No file found at", credentialsPath)
	}
	cfg, err := ini.Load(credentialsPath)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func writeCredentialsFile(cfg *ini.File) string {
	homeDirectory := getHomeDirectory()
	credentialsPath := filepath.Join(homeDirectory, ".aws", "credentials")
	cfg.SaveTo(credentialsPath)

	return credentialsPath
}

func getMfaSerialNumber() string {
	homeDirectory := getHomeDirectory()
	credentialsPath := filepath.Join(homeDirectory, ".aws", "config")
	_, err := os.Stat(credentialsPath)
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := ini.Load(credentialsPath)

	return cfg.Section("default").Key("mfa_serial").String()
}

func deleteCurrentIamKey(client *iam.IAM) {
	output, err := client.ListAccessKeys(&iam.ListAccessKeysInput{
		UserName: nil,
	})
	if err != nil {
		log.Fatal(err)
	}
	cfg := readCredentialsFile()
	usedKey := cfg.Section("default_original").Key("aws_access_key_id").String()
	if usedKey == "" {
		log.Fatal("failed to fetch aws_access_key_id from default_original section in ~/.aws/credentials")
	}
	for index := 0; index < len(output.AccessKeyMetadata); index++ {
		candidate := *output.AccessKeyMetadata[index].AccessKeyId
		if candidate != usedKey {
			continue
		}
		fmt.Println("deleting IAM key", candidate)
		_, err := client.DeleteAccessKey(&iam.DeleteAccessKeyInput{
			AccessKeyId: &candidate,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

// RotateAccessKeys ensures that the user has a single, new IAM access key
func RotateAccessKeys() string {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           "default",
	}))
	client := iam.New(sess)
	deleteCurrentIamKey(client)
	newKeyOutput, err := client.CreateAccessKey(&iam.CreateAccessKeyInput{})
	if err != nil {
		log.Fatal(err)
	}
	cfg := readCredentialsFile()
	fmt.Println("new IAM key is ", *newKeyOutput.AccessKey.AccessKeyId)
	cfg.Section("default_original").Key("aws_access_key_id").SetValue(*newKeyOutput.AccessKey.AccessKeyId)
	cfg.Section("default_original").Key("aws_secret_access_key").SetValue(*newKeyOutput.AccessKey.SecretAccessKey)
	location := writeCredentialsFile(cfg)

	return location
}

// UpdateSessionKeys uses the default_original profile to get new session keys
// for the default profile
func UpdateSessionKeys() string {
	var tokenCode string
	fmt.Print("Please enter mfa code: ")
	fmt.Scanln(&tokenCode)
	serialNumber := getMfaSerialNumber()
	sessionLifespan := int64(60 * 60 * 36)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           "default_original",
	}))
	client := sts.New(sess)
	output, err := client.GetSessionToken(&sts.GetSessionTokenInput{
		SerialNumber:    &serialNumber,
		TokenCode:       &tokenCode,
		DurationSeconds: &sessionLifespan,
	})
	if err != nil {
		log.Fatal(err)
	}
	cfg := readCredentialsFile()

	cfg.Section("default").Key("aws_access_key_id").SetValue(*output.Credentials.AccessKeyId)
	cfg.Section("default").Key("aws_secret_access_key").SetValue(*output.Credentials.SecretAccessKey)
	cfg.Section("default").Key("aws_session_token").SetValue(*output.Credentials.SessionToken)
	location := writeCredentialsFile(cfg)
	return location
}
