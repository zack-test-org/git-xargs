package cmd

// RunStats will be a stats-tracker class that keeps score of which repos were touched, which were considered for update, which had branches made, PRs made, which were missing workflows or contexts, or had out of date workflows syntax values, etc
type RunStats struct {
	RepoCount int
}

// IncrementRepoCount ups the number of total repos considered by this tool
func (r *RunStats) IncrementRepoCount() {
	r.RepoCount++
}
