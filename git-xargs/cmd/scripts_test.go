package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensure that passing missing or misnamed scripts to the VerifyScripts function results in proper filtering
func TestVerifyScriptsRejectsMissingScripts(t *testing.T) {
	allBadScriptNames := []string{"./scripts/vincent-freeman.sh"}

	filteredScriptCollection, verifyErr := VerifyScripts(allBadScriptNames)

	assert.Error(t, verifyErr)

	assert.Equal(t, len(filteredScriptCollection.Scripts), 0)
}

func TestVerifyScriptsFindsValidScriptByShortname(t *testing.T) {
	goodScriptName := []string{"./_testscripts/add-license.sh"}

	filteredScriptCollection, verifyErr := VerifyScripts(goodScriptName)

	assert.NoError(t, verifyErr)

	assert.Equal(t, len(filteredScriptCollection.Scripts), 1)
}

func TestVerifyScriptsFindsValidScripts(t *testing.T) {
	goodScriptName := []string{"./_testscripts/add-license.sh", "./_testscripts/test-ruby.rb", "./_testscripts/test-python.py"}

	filteredScriptCollection, verifyErr := VerifyScripts(goodScriptName)

	assert.NoError(t, verifyErr)

	assert.Equal(t, len(filteredScriptCollection.Scripts), 3)
}

func TestVerifyScriptRejectsEmptyScriptList(t *testing.T) {
	emptyScriptList := []string{}

	filteredScriptCollection, verifyErr := VerifyScripts(emptyScriptList)

	assert.Error(t, verifyErr)

	assert.Equal(t, len(filteredScriptCollection.Scripts), 0)
}

func TestFileThatIsNotExecutableThrowsError(t *testing.T) {

	notExecutableScriptList := []string{"./_testscripts/bad-perm.sh"}

	_, verifyErr := VerifyScripts(notExecutableScriptList)

	assert.EqualError(t, verifyErr, "All scripts must be chmod'd to be executable by at least their owner")

}
