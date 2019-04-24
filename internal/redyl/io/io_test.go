package io

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var fakeHome = filepath.Join("/", "tmp", "redyl", "io_test", ".aws")

func getFakeTokenCode() string {
	return "fake"
}

func getFakeHomeDirectory() string {
	return filepath.Join("/", "tmp", "redyl", "io_test")
}

func getFakeSessionKeys(p string, t string, s string) map[string]string {
	params := make(map[string]string)
	params["aws_access_key_id"] = "fake_session_key"
	params["aws_secret_access_key"] = "fake_session_secret_key"
	params["aws_session_token"] = "fake_session_token"
	return params
}

func fakeDeleteIamKey(a string, b string) {}

func fakeGetNewIamKey(a string) map[string]string {
	params := make(map[string]string)
	params["aws_access_key_id"] = "fake_iam_key"
	params["aws_secret_access_key"] = "fake_iam_secret_key"
	return params
}

type TestInput struct {
	config      string
	credentials string
	profile     string
}

var testtable = []struct {
	in     TestInput
	target string
}{
	{TestInput{config: "testdata/default_config", credentials: "testdata/default_credentials", profile: "default"}, "testdata/default_target"},
	{TestInput{config: "testdata/named_config", credentials: "testdata/named_credentials", profile: "named"}, "testdata/named_target"},
}


func TestEndToEnd(t *testing.T) {
	for _, tt := range testtable {
		t.Run(tt.in.profile, func(t *testing.T) {
			copyFile(tt.in.credentials, filepath.Join(fakeHome, "credentials"))
			copyFile(tt.in.config, filepath.Join(fakeHome, "config"))
			// TODO test that mock AWS functions are getting called with correct
			// arguments
			updater := SessionKeyUpdater{
				getTokenCode:     getFakeTokenCode,
				getHomeDirectory: getFakeHomeDirectory,
				getSessionKeys:   getFakeSessionKeys,
			}
			rotator := AccessKeyRotator{
				getHomeDirectory: getFakeHomeDirectory,
				deleteIamKey:     fakeDeleteIamKey,
				createIamKey:     fakeGetNewIamKey,
			}
			updater.update(tt.in.profile)
			location := rotator.rotate(tt.in.profile)
			actualContent, err := ioutil.ReadFile(location)
			if err != nil {
				panic(err)
			}
			targetContent, err := ioutil.ReadFile(tt.target)
			if err != nil {
				panic(err)
			}
			actual := string(actualContent)
			target := string(targetContent)
			if actual != target {
				t.Error("Expected ", target, ", got ", actual)
			}

		})
	}
}

func copyFile(infile string, outfile string) {
	b, err := ioutil.ReadFile(infile)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(outfile, b, 0644)
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	os.RemoveAll(fakeHome)
	os.MkdirAll(fakeHome, os.ModePerm)
	outCode := m.Run()
	os.RemoveAll(fakeHome)
	os.Exit(outCode)
}
