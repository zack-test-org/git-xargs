package keybase

import (
	"io/ioutil"

	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/gruntwork-cli/shell"
)

func DecodeSecret(cipherText string) (string, error) {
	tmpPath, err := tmpFile()
	if err != nil {
		return "", errors.WithStackTrace(err)
	}
	options := shell.NewShellOptions()
	options.SensitiveArgs = true
	err = shell.RunShellCommand(options, "keybase", "decrypt", "-m", cipherText, "-o", tmpPath)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadFile(tmpPath)
	if err != nil {
		return "", errors.WithStackTrace(err)
	}
	return string(data), nil
}

func tmpFile() (string, error) {
	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()
	return tmpfile.Name(), nil
}
