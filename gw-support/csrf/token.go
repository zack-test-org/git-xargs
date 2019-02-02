package csrf

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/gruntwork-cli/files"
	"github.com/mitchellh/go-homedir"
)

const (
	Username                 = "gwsupport"
	csrfTokenFile            = "csrf-token.txt"
	DefaultGWSupportPathName = ".gw-support"
)

// This is intentionally a var so it can be modified in tests

func ensureGWSupportDir(gwSupportPathName string) (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	if gwSupportPathName == "" {
		gwSupportPathName = DefaultGWSupportPathName
	}
	homePath := filepath.Join(home, gwSupportPathName)

	if !files.FileExists(homePath) {
		if err := os.MkdirAll(homePath, 0700); err != nil {
			return homePath, errors.WithStackTrace(err)
		}
		return homePath, nil
	}

	return homePath, nil
}

func getCsrfTokenPath(gwSupportPathName string) (string, error) {
	homePath, err := ensureGWSupportDir(gwSupportPathName)
	if err != nil {
		return "", err
	}
	csrfTokenPath := filepath.Join(homePath, csrfTokenFile)
	return csrfTokenPath, nil
}

func CreateCsrfToken(gwSupportPathName string) (string, error) {
	csrfTokenPath, err := getCsrfTokenPath(gwSupportPathName)
	if err != nil {
		return "", err
	}

	tokenStr := uuid.New().String()
	err = ioutil.WriteFile(csrfTokenPath, []byte(tokenStr), 0600)
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	return tokenStr, nil
}

func DeleteCsrfToken(gwSupportPathName string) error {
	csrfTokenPath, err := getCsrfTokenPath(gwSupportPathName)
	if err != nil {
		return err
	}

	if files.FileExists(csrfTokenPath) {
		if err := os.Remove(csrfTokenPath); err != nil {
			return errors.WithStackTrace(err)
		}
	}
	return nil
}

func GetOrCreateCsrfToken(gwSupportPathName string) (string, error) {
	csrfTokenPath, err := getCsrfTokenPath(gwSupportPathName)
	if err != nil {
		return "", err
	}

	if !files.FileExists(csrfTokenPath) {
		return CreateCsrfToken(gwSupportPathName)
	}

	token, err := ioutil.ReadFile(csrfTokenPath)
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	return string(token), nil
}
