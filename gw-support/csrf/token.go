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
	Username      = "gwsupport"
	csrfTokenFile = "csrf-token.txt"
)

// This is intentionally a var so it can be modified in tests
var gwSupportPathRelativeHome = ".gw-support"

func ensureGWSupportDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	homePath := filepath.Join(home, gwSupportPathRelativeHome)

	if !files.FileExists(homePath) {
		if err := os.MkdirAll(homePath, 0700); err != nil {
			return homePath, errors.WithStackTrace(err)
		}
		return homePath, nil
	}

	return homePath, nil
}

func getCsrfTokenPath() (string, error) {
	homePath, err := ensureGWSupportDir()
	if err != nil {
		return "", err
	}
	csrfTokenPath := filepath.Join(homePath, csrfTokenFile)
	return csrfTokenPath, nil
}

func CreateCsrfToken() (string, error) {
	csrfTokenPath, err := getCsrfTokenPath()
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

func DeleteCsrfToken() error {
	csrfTokenPath, err := getCsrfTokenPath()
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

func GetOrCreateCsrfToken() (string, error) {
	csrfTokenPath, err := getCsrfTokenPath()
	if err != nil {
		return "", err
	}

	if !files.FileExists(csrfTokenPath) {
		return CreateCsrfToken()
	}

	token, err := ioutil.ReadFile(csrfTokenPath)
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	return string(token), nil
}
