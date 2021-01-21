package cmd

import (
	"os/exec"

	"github.com/sirupsen/logrus"
)

// handles required dependency lookups on startup
// Checks if a given program is installed locally - similar to a `which` command
func dependencyInstalled(dep string) bool {
	_, err := exec.LookPath(dep)
	return err == nil
}

// Dependency represents a third party binary that must be installed on the operator's system in order for them to use this tool
type Dependency struct {
	Name string
	URL  string
}

// MustHaveDependenciesInstalled accepts a slice of dependencies, and FREAKS OUT if any of them are missing
func verifyDependenciesInstalled(deps []Dependency) (bool, []Dependency) {
	var missingDeps []Dependency

	for _, d := range deps {

		if !dependencyInstalled(d.Name) {
			missingDeps = append(missingDeps, d)
		}
	}
	return len(missingDeps) == 0, missingDeps
}

// Sanity check that user has at least one ssh identity loaded (since git auth is done via ssh - so their clones would fail)
func EnsureSSHAgentHasIdentitiesLoaded() bool {
	out, err := exec.Command("ssh-add", "-l").Output()
	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
		}).Debug("Could not check ssh-add for loaded identities")

		return false
	}

	return string(out) != "The agent has no identities."

}
