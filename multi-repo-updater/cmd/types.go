package cmd

// AllowedRepo represents a single repository under a Github organization that this tool may operate on
type AllowedRepo struct {
	Organization string
	Name         string
}
