package cmd

// AllowedRepo represents a single repository under a Github organization that this tool may operate on
type AllowedRepo struct {
	Organization string `header:"Organization name"`
	Name         string `header:"URL"`
}

// ReducedRepo is a simplified form of the github.Repository struct
type ReducedRepo struct {
	Name string `header:"Repo name"`
	URL  string `header:"Repo url"`
}

// OpenedPullRequest is a simple two column representation of the repo name and its PR url
type PullRequest struct {
	Repo string `header:"Repo name"`
	URL  string `header:"PR URL"`
}

// Script represents a single shell script to be run against a repo
type Script struct {
	Path string
}

// Script collection contains a slice of scripts that are to be executed against the local copies of each targeted repo
type ScriptCollection struct {
	Scripts []Script
}

// Add accepts a single script and appends it to the internal slice of scripts to run within the script collection
func (sc *ScriptCollection) Add(s Script) {
	sc.Scripts = append(sc.Scripts, s)
}
