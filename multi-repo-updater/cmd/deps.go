package cmd

import (
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// handles required dependency lookups on startup
// Checks if a given program is installed locally - similar to a `which` command

func dependencyInstalled(dep string) bool {
	_, err := exec.LookPath(dep)
	return err == nil
}

// A given third party binary that must be installed on the operator's system in order for them to use this tool
type Dependency struct {
	Name string
	URL  string
}

// Accepts a slice of dependencies, and FREAKS OUT if any of them are missing
func MustHaveDependenciesInstalled(deps []Dependency) {

	for _, d := range deps {

		if !dependencyInstalled(d.Name) {
			log.WithFields(logrus.Fields{
				"Dependency":         d.Name,
				"Install / info URL": d.URL,
			}).Debug("Missing dependency. Please install it before using this tool")
			os.Exit(1)
		}
	}
}
