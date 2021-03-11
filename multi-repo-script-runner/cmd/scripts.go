package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// IsExecutableByOwner returns true if a given file's mode means its owner can exexecute it, and false if they cannot
func IsExecutableByOwner(mode os.FileMode) bool {
	return mode&0100 != 0
}

// VerifyScripts runs a sanity check against each supplied script, ensuring that it exists and can be read
// It then packages all the supplied scripts into a ScriptCollection struct so that all scripts are available with their full paths when needed during execution
func VerifyScripts(scriptNames []string, scriptsPath string) (ScriptCollection, error) {
	// Create a ScriptsCollection that will hold all the scripts that should be run against the repos
	sc := ScriptCollection{}

	// Sanity check that the user passed scripts to the tool at all
	if len(scriptNames) == 0 {
		return sc, errors.New("You must provide at least one valid script that exists in the ./scripts directory to run against selected repos")
	}

	var basePath string

	// If the scriptsPath was passed, prefer it over the path of the executable itself. This assists with testing
	if scriptsPath != "" {
		basePath = scriptsPath
	} else {
		// Get path to the executable itself
		ex, err := os.Executable()
		if err != nil {
			return sc, err
		}
		basePath = filepath.Join(filepath.Dir(ex), "scripts")
	}

	// Ensure that every script can be read by the tool, ensuring there were no naming or permissions issues
	for _, scriptName := range scriptNames {
		// Build complete path to the script
		scriptPath := filepath.Join(basePath, scriptName)

		file, openErr := os.Open(scriptPath)
		if openErr != nil {
			log.WithFields(logrus.Fields{
				"Error":       openErr,
				"Script name": scriptName,
				"Script path": scriptPath,
			}).Debug("Every target script must exist in ./scripts and be readable and executable. Cannot run this script!")
			return sc, openErr
		}

		fileInfo, statErr := file.Stat()

		if statErr != nil {
			log.WithFields(logrus.Fields{
				"Error":       statErr,
				"Script path": scriptPath,
			}).Debug("Error getting file info")
			return sc, statErr
		}

		// Sanity check that script is executable by its owner
		if !IsExecutableByOwner(fileInfo.Mode()) {
			log.WithFields(logrus.Fields{
				"File":      scriptPath,
				"File Mode": fileInfo.Mode(),
			}).Debug("File is not executable by owner")
			return sc, errors.New("All scripts must be chmod'd to be executable by at least their owner")

		}

		// Script passed sanity check - we were able to find and open it
		// Package it as a script type and add it to the ScriptCollection
		s := Script{
			Path: scriptPath,
		}

		sc.Add(s)
	}

	return sc, nil
}
