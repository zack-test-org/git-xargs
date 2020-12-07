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
