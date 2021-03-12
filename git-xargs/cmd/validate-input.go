package cmd

// Sanity check that user has provided one valid method for selecting repos to operate on
func ensureValidOptionsPassed(allowedReposFile, GithubOrg string) bool {
	return allowedReposFile != "" || GithubOrg != ""
}
