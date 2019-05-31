package io

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/pbs/redyl/internal/redyl/aws"

	ini "gopkg.in/ini.v1"
)

func getUserHomeDirectory() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func readAWSIniFile(filename string, home string) *ini.File {
	location := filepath.Join(home, ".aws", filename)
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

func writeAWSIniFile(filename string, cfg *ini.File, home string) string {
	location := filepath.Join(home, ".aws", filename)
	cfg.SaveTo(location)

	return location
}

func updateCredentials(section string, parameters map[string]string, home string) string {
	creds := readAWSIniFile("credentials", home)
	for k, v := range parameters {
		creds.Section(section).Key(k).SetValue(v)
	}
	location := writeAWSIniFile("credentials", creds, home)

	return location
}

func getMfaSerialNumber(profile string, home string) string {
	config := readAWSIniFile("config", home)
	val := config.Section(profile).Key("mfa_serial").String()
	if val == "" {
		log.Fatal("failed to fetch mfa_serial from ", profile, "section in ~/.aws/config")
	}

	return val
}

func getCurrentIamKey(profile string, home string) string {
	creds := readAWSIniFile("credentials", home)
	key := creds.Section(profile).Key("aws_access_key_id").String()
	if key == "" {
		log.Fatal("failed to fetch aws_access_key_id from ", profile, " section in ~/.aws/credentials")
	}
	return key
}

func getTokenCodeFromMfa() string {
	var tokenCode string
	fmt.Print("Please enter mfa code: ")
	fmt.Scanln(&tokenCode)

	return tokenCode
}

// SessionKeyUpdater updates session keys
type SessionKeyUpdater struct {
	getTokenCode     func() string
	getHomeDirectory func() string
	getSessionKeys   func(string, string, string) map[string]string
}

func (s SessionKeyUpdater) update(profile string) string {
	home := s.getHomeDirectory()
	token := s.getTokenCode()
	serial := getMfaSerialNumber(profile, home)
	params := s.getSessionKeys(profile+"_original", token, serial)
	location := updateCredentials(profile, params, home)
	if profile != "default" {
		location = updateCredentials("default", params, home)
	}

	return location
}

// AccessKeyRotator updates access keys
type AccessKeyRotator struct {
	getHomeDirectory func() string
	deleteIamKey     func(string, string)
	createIamKey     func(string) map[string]string
}

func (a AccessKeyRotator) rotate(profile string) string {
	home := a.getHomeDirectory()
	profileOriginal := profile + "_original"
	key := getCurrentIamKey(profileOriginal, home)
	a.deleteIamKey(profile, key)
	params := a.createIamKey(profile)
	location := updateCredentials(profileOriginal, params, home)

	return location
}

// UpdateSessionKeys uses the original profile to get new session keys
// for the profile
func UpdateSessionKeys(profile string) string {
	updater := SessionKeyUpdater{
		getTokenCode:     getTokenCodeFromMfa,
		getHomeDirectory: getUserHomeDirectory,
		getSessionKeys:   aws.GetSessionKeys,
	}
	location := updater.update(profile)
	return location
}

// RotateAccessKeys uses the profile to get new access keys for the
// original profile. In the process, it deletes the current original access key
func RotateAccessKeys(profile string) string {
	rotator := AccessKeyRotator{
		getHomeDirectory: getUserHomeDirectory,
		deleteIamKey:     aws.DeleteIamKey,
		createIamKey:     aws.GetNewIamKey,
	}
	location := rotator.rotate(profile)

	return location
}
