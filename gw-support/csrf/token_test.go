package csrf

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/gruntwork-cli/files"
	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests in this file manipulate global variables, so cannot be run in parallel or they will conflict. However,
// these tests are tiny so speed shouldn't be an issue.

func TestEnsureGWSupportDirCreatesDirIfNotExists(t *testing.T) {
	// Override home path for testing purposes
	gwSupportPathRelativeHome = fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t)

	// If dir exists, remove it
	if files.FileExists(expectedGWSupportPath) {
		require.NoError(t, os.RemoveAll(expectedGWSupportPath))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	path, err := ensureGWSupportDir()
	require.NoError(t, err)
	assert.Equal(t, path, expectedGWSupportPath)
	assert.True(t, files.FileExists(path))
	assert.True(t, files.IsDir(path))
}

func TestEnsureGWSupportDirWorksIfDirExists(t *testing.T) {
	// Override home path for testing purposes
	gwSupportPathRelativeHome = fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t)

	// If dir does not exist, make it
	if !files.FileExists(expectedGWSupportPath) {
		require.NoError(t, os.MkdirAll(expectedGWSupportPath, 0700))
		require.True(t, files.FileExists(expectedGWSupportPath))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	path, err := ensureGWSupportDir()
	require.NoError(t, err)
	assert.Equal(t, path, expectedGWSupportPath)
	assert.True(t, files.FileExists(path))
	assert.True(t, files.IsDir(path))
}

func TestCreateCsrfTokenCreatesATokenWhenFileDoesNotExist(t *testing.T) {
	// Override home path for testing purposes
	gwSupportPathRelativeHome = fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t)
	csrfPath := filepath.Join(expectedGWSupportPath, csrfTokenFile)

	// If csrf path exists, remove it
	if files.FileExists(csrfPath) {
		require.NoError(t, os.RemoveAll(expectedGWSupportPath))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	tokenStr, err := CreateCsrfToken()
	require.NoError(t, err)
	data, err := ioutil.ReadFile(csrfPath)
	require.NoError(t, err)
	assert.Equal(t, string(data), tokenStr)
}

func TestCreateCsrfTokenCreatesANewTokenWhenFileExist(t *testing.T) {
	// Override home path for testing purposes
	gwSupportPathRelativeHome = fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t)
	csrfPath := filepath.Join(expectedGWSupportPath, csrfTokenFile)

	// If csrf path does not exist, create it
	if !files.FileExists(csrfPath) {
		if !files.FileExists(expectedGWSupportPath) {
			require.NoError(t, os.MkdirAll(expectedGWSupportPath, 0700))
		}
		require.NoError(t, ioutil.WriteFile(csrfPath, []byte{}, 0700))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	tokenStr, err := CreateCsrfToken()
	require.NoError(t, err)
	data, err := ioutil.ReadFile(csrfPath)
	require.NoError(t, err)
	assert.Equal(t, string(data), tokenStr)
}

func TestDeleteCsrfTokenDeletesFile(t *testing.T) {
	// Override home path for testing purposes
	gwSupportPathRelativeHome = fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t)
	csrfPath := filepath.Join(expectedGWSupportPath, csrfTokenFile)

	// If csrf path does not exist, create it
	if !files.FileExists(csrfPath) {
		if !files.FileExists(expectedGWSupportPath) {
			require.NoError(t, os.MkdirAll(expectedGWSupportPath, 0700))
		}
		require.NoError(t, ioutil.WriteFile(csrfPath, []byte{}, 0700))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	require.NoError(t, DeleteCsrfToken())
	assert.False(t, files.FileExists(csrfPath))
}

func TestDeleteCsrfTokenSucceedsIfDirDoesNotExist(t *testing.T) {
	// Override home path for testing purposes
	gwSupportPathRelativeHome = fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t)
	csrfPath := filepath.Join(expectedGWSupportPath, csrfTokenFile)

	// If csrf path exists, remove it
	if files.FileExists(csrfPath) {
		require.NoError(t, os.RemoveAll(expectedGWSupportPath))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	require.NoError(t, DeleteCsrfToken())
	assert.False(t, files.FileExists(csrfPath))
}

func TestGetOrCreateCsrfTokenCreatesTokenIfNotExist(t *testing.T) {
	// Override home path for testing purposes
	gwSupportPathRelativeHome = fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t)
	csrfPath := filepath.Join(expectedGWSupportPath, csrfTokenFile)

	// If csrf path exists, remove it
	if files.FileExists(csrfPath) {
		require.NoError(t, os.RemoveAll(expectedGWSupportPath))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	tokenStr, err := GetOrCreateCsrfToken()
	require.NoError(t, err)
	data, err := ioutil.ReadFile(csrfPath)
	require.NoError(t, err)
	assert.Equal(t, string(data), tokenStr)
}

func TestGetOrCreateCsrfTokenReturnsTokenIfExist(t *testing.T) {
	// Override home path for testing purposes
	gwSupportPathRelativeHome = fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t)
	csrfPath := filepath.Join(expectedGWSupportPath, csrfTokenFile)

	// If csrf path does not exist, create it
	if !files.FileExists(csrfPath) {
		if !files.FileExists(expectedGWSupportPath) {
			require.NoError(t, os.MkdirAll(expectedGWSupportPath, 0700))
		}
		require.NoError(t, ioutil.WriteFile(csrfPath, []byte("foo"), 0700))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	tokenStr, err := GetOrCreateCsrfToken()
	require.NoError(t, err)
	assert.Equal(t, tokenStr, "foo")
	data, err := ioutil.ReadFile(csrfPath)
	require.NoError(t, err)
	assert.Equal(t, string(data), tokenStr)
}

func expectedDir(t *testing.T) string {
	home, err := homedir.Dir()
	require.NoError(t, err)
	return filepath.Join(home, gwSupportPathRelativeHome)
}
