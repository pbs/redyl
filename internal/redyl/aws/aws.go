package aws

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	iam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
)

func getAWSSession(profile string) *session.Session {
	vars := []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN"}
	for _, v := range vars {
		if os.Getenv(v) != "" {
			log.Fatal("please unset the environment variable ", v)
		}
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
	}))

	return sess
}

// DeleteIamKey deletes an IAM key from the user's account
func DeleteIamKey(profile string, key string) {
	sess := getAWSSession(profile)
	client := iam.New(sess)
	output, err := client.ListAccessKeys(&iam.ListAccessKeysInput{
		UserName: nil,
	})
	if err != nil {
		log.Fatal(err)
	}

	for index := 0; index < len(output.AccessKeyMetadata); index++ {
		candidate := *output.AccessKeyMetadata[index].AccessKeyId
		if candidate != key {
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

// GetNewIamKey creates a new IAM key in the user's account
func GetNewIamKey(profile string) map[string]string {
	sess := getAWSSession(profile)
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

// GetSessionKeys fetches short-lived session keys
func GetSessionKeys(profile string, tokenCode string, serialNumber string) map[string]string {
	sessionLifespan := int64(60 * 60 * 36)
	sess := getAWSSession(profile)
	client := sts.New(sess)
	output, err := client.GetSessionToken(&sts.GetSessionTokenInput{
		SerialNumber:    &serialNumber,
		TokenCode:       &tokenCode,
		DurationSeconds: &sessionLifespan,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "InvalidClientTokenId":
				log.Fatal("check your security credentials at https://console.aws.amazon.com/iam/home to ensure that your access key is active")
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
