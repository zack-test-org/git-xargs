package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensure that passing missing or misnamed scripts to the VerifyScripts function results in proper filtering
func TestVerifyScriptsRejectsMissingScripts(t *testing.T) {
	allBadScriptNames := []string{"vincent-freeman", "booker-dewitt.sh", "im-not-really-a-script"}

	filteredScriptCollection, verifyErr := VerifyScripts(allBadScriptNames, "_testscripts")

	assert.Error(t, verifyErr)

	assert.Equal(t, len(filteredScriptCollection.Scripts), 0)
}

func TestVerifyScriptsFindsValidScriptByShortname(t *testing.T) {
	goodScriptName := []string{"add-license"}

	filteredScriptCollection, verifyErr := VerifyScripts(goodScriptName, "_testscripts")

	assert.NoError(t, verifyErr)

	assert.Equal(t, len(filteredScriptCollection.Scripts), 1)

	assert.Equal(t, filteredScriptCollection.Scripts[0].Path, "_testscripts/add-license.sh")

}

func TestVerifyScriptsFindsValidScriptWithShExtension(t *testing.T) {
	goodScriptName := []string{"add-license.sh"}

	filteredScriptCollection, verifyErr := VerifyScripts(goodScriptName, "_testscripts")

	assert.NoError(t, verifyErr)

	assert.Equal(t, len(filteredScriptCollection.Scripts), 1)

	assert.Equal(t, filteredScriptCollection.Scripts[0].Path, "_testscripts/add-license.sh")

}

func TestVerifyScriptRejectsEmptyScriptList(t *testing.T) {
	emptyScriptList := []string{}

	filteredScriptCollection, verifyErr := VerifyScripts(emptyScriptList, "_testscripts")

	assert.Error(t, verifyErr)

	assert.Equal(t, len(filteredScriptCollection.Scripts), 0)

}

func TestFileThatIsNotExecutableThrowsError(t *testing.T) {

	notExecutableScriptList := []string{"bad-perm"}

	_, verifyErr := VerifyScripts(notExecutableScriptList, "_testscripts")

	assert.EqualError(t, verifyErr, "All scripts must be chmod'd to be executable by at least their owner")

}
