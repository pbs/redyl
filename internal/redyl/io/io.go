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

	return config.Section(profile).Key("mfa_serial").String()
}

func getCurrentIamKey(profile string, home string) string {
	creds := readAWSIniFile("credentials", home)
	key := creds.Section(profile).Key("aws_access_key_id").String()
	if key == "" {
		log.Fatal("failed to fetch aws_access_key_id from ", profile, "section in ~/.aws/credentials")
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

func (s SessionKeyUpdater) update() string {
	home := s.getHomeDirectory()
	token := s.getTokenCode()
	serial := getMfaSerialNumber("default", home)
	params := s.getSessionKeys("default_original", token, serial)
	location := updateCredentials("default", params, home)

	return location
}

// AccessKeyRotator updates session keys
type AccessKeyRotator struct {
	getHomeDirectory func() string
	deleteIamKey     func(string, string)
	createIamKey     func(string) map[string]string
}

func (a AccessKeyRotator) rotate() string {
	home := a.getHomeDirectory()
	key := getCurrentIamKey("default_original", home)
	a.deleteIamKey("default_original", key)
	params := a.createIamKey("default")
	location := updateCredentials("default_original", params, home)

	return location
}

// UpdateSessionKeys uses the default_original profile to get new session keys
// for the default profile
func UpdateSessionKeys() string {
	updater := SessionKeyUpdater{
		getTokenCode:     getTokenCodeFromMfa,
		getHomeDirectory: getUserHomeDirectory,
		getSessionKeys:   aws.GetSessionKeys,
	}
	location := updater.update()
	return location
}

// RotateAccessKeys uses the default profile to get new access keys for the
// default_original profile. In the process, it deletes the current default_original access key
func RotateAccessKeys() string {
	rotator := AccessKeyRotator{
		getHomeDirectory: getUserHomeDirectory,
		deleteIamKey:     aws.DeleteIamKey,
		createIamKey:     aws.GetNewIamKey,
	}
	location := rotator.rotate()

	return location
}
