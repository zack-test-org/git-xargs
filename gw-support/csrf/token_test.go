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

func TestEnsureGWSupportDirCreatesDirIfNotExists(t *testing.T) {
	t.Parallel()

	gwSupportPathRelativeHome := fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t, gwSupportPathRelativeHome)

	// If dir exists, remove it
	if files.FileExists(expectedGWSupportPath) {
		require.NoError(t, os.RemoveAll(expectedGWSupportPath))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	path, err := ensureGWSupportDir(gwSupportPathRelativeHome)
	require.NoError(t, err)
	assert.Equal(t, path, expectedGWSupportPath)
	assert.True(t, files.FileExists(path))
	assert.True(t, files.IsDir(path))
}

func TestEnsureGWSupportDirWorksIfDirExists(t *testing.T) {
	t.Parallel()

	gwSupportPathRelativeHome := fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t, gwSupportPathRelativeHome)

	// If dir does not exist, make it
	if !files.FileExists(expectedGWSupportPath) {
		require.NoError(t, os.MkdirAll(expectedGWSupportPath, 0700))
		require.True(t, files.FileExists(expectedGWSupportPath))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	path, err := ensureGWSupportDir(gwSupportPathRelativeHome)
	require.NoError(t, err)
	assert.Equal(t, path, expectedGWSupportPath)
	assert.True(t, files.FileExists(path))
	assert.True(t, files.IsDir(path))
}

func TestCreateCsrfTokenCreatesATokenWhenFileDoesNotExist(t *testing.T) {
	t.Parallel()

	gwSupportPathRelativeHome := fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t, gwSupportPathRelativeHome)
	csrfPath := filepath.Join(expectedGWSupportPath, csrfTokenFile)

	// If csrf path exists, remove it
	if files.FileExists(csrfPath) {
		require.NoError(t, os.RemoveAll(expectedGWSupportPath))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	tokenStr, err := CreateCsrfToken(gwSupportPathRelativeHome)
	require.NoError(t, err)
	data, err := ioutil.ReadFile(csrfPath)
	require.NoError(t, err)
	assert.Equal(t, string(data), tokenStr)
}

func TestCreateCsrfTokenCreatesANewTokenWhenFileExist(t *testing.T) {
	t.Parallel()

	gwSupportPathRelativeHome := fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t, gwSupportPathRelativeHome)
	csrfPath := filepath.Join(expectedGWSupportPath, csrfTokenFile)

	// If csrf path does not exist, create it
	if !files.FileExists(csrfPath) {
		if !files.FileExists(expectedGWSupportPath) {
			require.NoError(t, os.MkdirAll(expectedGWSupportPath, 0700))
		}
		require.NoError(t, ioutil.WriteFile(csrfPath, []byte{}, 0700))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	tokenStr, err := CreateCsrfToken(gwSupportPathRelativeHome)
	require.NoError(t, err)
	data, err := ioutil.ReadFile(csrfPath)
	require.NoError(t, err)
	assert.Equal(t, string(data), tokenStr)
}

func TestDeleteCsrfTokenDeletesFile(t *testing.T) {
	t.Parallel()

	gwSupportPathRelativeHome := fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t, gwSupportPathRelativeHome)
	csrfPath := filepath.Join(expectedGWSupportPath, csrfTokenFile)

	// If csrf path does not exist, create it
	if !files.FileExists(csrfPath) {
		if !files.FileExists(expectedGWSupportPath) {
			require.NoError(t, os.MkdirAll(expectedGWSupportPath, 0700))
		}
		require.NoError(t, ioutil.WriteFile(csrfPath, []byte{}, 0700))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	require.NoError(t, DeleteCsrfToken(gwSupportPathRelativeHome))
	assert.False(t, files.FileExists(csrfPath))
}

func TestDeleteCsrfTokenSucceedsIfDirDoesNotExist(t *testing.T) {
	t.Parallel()

	gwSupportPathRelativeHome := fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t, gwSupportPathRelativeHome)
	csrfPath := filepath.Join(expectedGWSupportPath, csrfTokenFile)

	// If csrf path exists, remove it
	if files.FileExists(csrfPath) {
		require.NoError(t, os.RemoveAll(expectedGWSupportPath))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	require.NoError(t, DeleteCsrfToken(gwSupportPathRelativeHome))
	assert.False(t, files.FileExists(csrfPath))
}

func TestGetOrCreateCsrfTokenCreatesTokenIfNotExist(t *testing.T) {
	t.Parallel()

	gwSupportPathRelativeHome := fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t, gwSupportPathRelativeHome)
	csrfPath := filepath.Join(expectedGWSupportPath, csrfTokenFile)

	// If csrf path exists, remove it
	if files.FileExists(csrfPath) {
		require.NoError(t, os.RemoveAll(expectedGWSupportPath))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	tokenStr, err := GetOrCreateCsrfToken(gwSupportPathRelativeHome)
	require.NoError(t, err)
	data, err := ioutil.ReadFile(csrfPath)
	require.NoError(t, err)
	assert.Equal(t, string(data), tokenStr)
}

func TestGetOrCreateCsrfTokenReturnsTokenIfExist(t *testing.T) {
	t.Parallel()

	gwSupportPathRelativeHome := fmt.Sprintf(".gw-support-test-%s", t.Name())
	expectedGWSupportPath := expectedDir(t, gwSupportPathRelativeHome)
	csrfPath := filepath.Join(expectedGWSupportPath, csrfTokenFile)

	// If csrf path does not exist, create it
	if !files.FileExists(csrfPath) {
		if !files.FileExists(expectedGWSupportPath) {
			require.NoError(t, os.MkdirAll(expectedGWSupportPath, 0700))
		}
		require.NoError(t, ioutil.WriteFile(csrfPath, []byte("foo"), 0700))
	}
	defer os.RemoveAll(expectedGWSupportPath)

	tokenStr, err := GetOrCreateCsrfToken(gwSupportPathRelativeHome)
	require.NoError(t, err)
	assert.Equal(t, tokenStr, "foo")
	data, err := ioutil.ReadFile(csrfPath)
	require.NoError(t, err)
	assert.Equal(t, string(data), tokenStr)
}

func expectedDir(t *testing.T, gwSupportPathRelativeHome string) string {
	home, err := homedir.Dir()
	require.NoError(t, err)
	return filepath.Join(home, gwSupportPathRelativeHome)
}
