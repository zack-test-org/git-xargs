package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// IsExecutableByOwner returns true if a given file's mode means its owner can exexecute it, and false if they cannot
func IsExecutableByOwner(mode os.FileMode) bool {
	return mode&0100 != 0
}

// VerifyScripts runs a sanity check against each supplied script, ensuring that it exists and can be read
// It then packages all the supplied scripts into a ScriptCollection struct so that all scripts are available with their full paths when needed during execution
func VerifyScripts(scriptPaths []string) (ScriptCollection, error) {
	// Create a ScriptsCollection that will hold all the scripts that should be run against the repos
	sc := ScriptCollection{}

	// Sanity check that the user passed scripts to the tool at all
	if len(scriptPaths) == 0 {
		return sc, errors.New("You must provide at least one valid script path that exists on this system and is executable")
	}

	// Ensure that every script can be read by the tool, ensuring there were no naming or permissions issues
	for _, scriptPath := range scriptPaths {
		scriptPath = strings.TrimSpace(scriptPath)
		// Check if relative path was passed for script and build it into an absolute path
		if !filepath.IsAbs(scriptPath) {
			abs, absErr := filepath.Abs(scriptPath)
			if absErr != nil {
				log.WithFields(logrus.Fields{
					"Error":       absErr,
					"Script path": scriptPath,
				}).Debug("Could not convert relative script path to absolute")
				return sc, absErr
			}
			log.WithFields(logrus.Fields{
				"Original script path": scriptPath,
				"Absolute script path": abs,
			}).Debug("Converted relative path script beginning with ./ to absolute path")
			scriptPath = abs
		}

		file, openErr := os.Open(scriptPath)
		if openErr != nil {
			log.WithFields(logrus.Fields{
				"Error":       openErr,
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
