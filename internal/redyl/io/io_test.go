package io

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

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

func TestEndToEnd(t *testing.T) {
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
	updater.update()
	location := rotator.rotate()
	actualContent, err := ioutil.ReadFile(location)
	if err != nil {
		panic(err)
	}
	targetContent, err := ioutil.ReadFile("testdata/target_credentials")
	if err != nil {
		panic(err)
	}
	actual := string(actualContent)
	target := string(targetContent)
	if actual != target {
		t.Error("Expected ", target, ", got ", actual)
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
	fakeHome := filepath.Join("/", "tmp", "redyl", "io_test", ".aws")
	os.RemoveAll(fakeHome)
	os.MkdirAll(fakeHome, os.ModePerm)
	copyFile("testdata/nice_credentials", filepath.Join(fakeHome, "credentials"))
	copyFile("testdata/nice_config", filepath.Join(fakeHome, "config"))
	outCode := m.Run()
	os.RemoveAll(fakeHome)
	os.Exit(outCode)
}
