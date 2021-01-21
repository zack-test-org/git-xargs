package cmd

import (
	"github.com/google/go-github/v32/github"
)

// Event is a generic tracking ocurrence that RunStats manages
type Event string

const (
	// DryRunSet denotes a repo will not have any file changes, branches made or PRs opened because the dry-run flag was set to true
	DryRunSet Event = "dry-run-set-no-changes-made"
	// ReposSelected denotes all the repositories that were targeted for processing by this tool AFTER filtering was applied to determine valid repos
	ReposSelected Event = "repos-selected-pre-processing"
	// TargetBranchNotFound denotes the special branch used by this tool to make changes on was not found on lookup, suggesting it should be created
	TargetBranchNotFound Event = "target-branch-not-found"
	// TargetBranchAlreadyExists denotes the special branch used by this tool was already found (so it was likely already created by a previous run)
	TargetBranchAlreadyExists Event = "target-branch-already-exists"
	// TargetBranchLookupErr denotes an issue performing the lookup via Github API for the target branch - an API call failure
	TargetBranchLookupErr Event = "target-branch-lookup-err"
	// TargetBranchSuccessfullyCreated denotes a repo for which the target branch was created via Github API call
	TargetBranchSuccessfullyCreated Event = "target-branch-successfully-created"
	// FetchedViaGithubAPI denotes a repo was successfully listed by calling the Github API
	FetchedViaGithubAPI Event = "fetch-via-github-api"
	// RepoSuccessfullyCloned denotes a repo that was able to be cloned to the local filesystem of the operator's machine
	RepoSuccessfullyCloned Event = "repo-successfully-cloned"
	// RepoFailedToClone denotes that for whatever reason we were unable to clone the repo to the local system
	RepoFailedToClone Event = "repo-failed-to-clone"
	// BranchCheckoutFailed denotes a failure to checkout a new tool specific branch in the given repo
	BranchCheckoutFailed Event = "branch-checkout-failed"
	// GetHeadRefFailed denotes a repo for which the HEAD git reference could not be obtained
	GetHeadRefFailed Event = "get-head-ref-failed"
	// ScriptErrorOccurredDuringExecution denotes a repo for which at least one script raised an error during execution
	ScriptErrorOcurredDuringExecution Event = "script-error-during-execution"
	// WorktreeStatusCheckFailed denotes a repo whose git status command failed post script execution
	WorktreeStatusCheckFailed Event = "worktree-status-check-failed"
	// WorktreeStatusDirty denotes repos that had local file changes following execution of all their targeted scripts
	WorktreeStatusDirty Event = "worktree-status-dirty"
	// WorktreeStatusClean denotes a repo that did not have any local file changes following script execution
	WorktreeStatusClean Event = "worktree-status-clean"
	// WorktreeAddFileFailed denotes a failure to add at least one file to the git stage following script execution
	WorktreeAddFileFailed Event = "worktree-add-file-failed"
	// CommitChangesFailed denotes an error git committing our file changes to the local repo
	CommitChangesFailed Event = "commit-changes-failed"
	// PushBranchFailed denotes a repo whose new tool-specific branch could not be pushed to remote origin
	PushBranchFailed Event = "push-branch-failed"
	// PushBranchSkipped denotes a repo whose local branch was not pushed due to the --dry-run flag being set
	PushBranchSkipped Event = "push-branch-skipped"
	// RepoNotExists denotes a repo + org combo that was supplied via file but could not be successfully looked up via the Github API (returned a 404)
	RepoNotExists Event = "repo-not-exists"
	// PullRequestOpenErr denotes a repo whose pull request containing config changes could not be made successfully
	PullRequestOpenErr Event = "pull-request-open-error"
)

// AnnotatedEvent is used in printing the final report. It contains the info to print a section's table - both it's Event for looking up the tagged repos, and the human-legible description for printing above the table
type AnnotatedEvent struct {
	Event       Event
	Description string
}

var allEvents = []AnnotatedEvent{
	{Event: FetchedViaGithubAPI, Description: "Repos successfully fetched via Github API"},
	{Event: DryRunSet, Description: "Repos that were not modified in any way because this was a dry-run"},
	{Event: ReposSelected, Description: "All repos that were targeted for processing AFTER filtering missing / malformed repos"},
	{Event: TargetBranchNotFound, Description: "Repos whose target branch was not found"},
	{Event: TargetBranchAlreadyExists, Description: "Repos whose target branch already existed"},
	{Event: TargetBranchLookupErr, Description: "Repos whose target branches could not be looked up due to an API error"},
	{Event: RepoSuccessfullyCloned, Description: "Repos that were successfully cloned to the local filesystem"},
	{Event: RepoFailedToClone, Description: "Repos that were unable to be cloned to the local filesystem"},
	{Event: BranchCheckoutFailed, Description: "Repos for which checking out a new tool-specific branch failed"},
	{Event: GetHeadRefFailed, Description: "Repos for which the HEAD git reference could not be obtained"},
	{Event: ScriptErrorOcurredDuringExecution, Description: "Repos for which at least one script raised an error during execution"},
	{Event: WorktreeStatusCheckFailed, Description: "Repos for which the git status command failed following script execution"},
	{Event: WorktreeStatusDirty, Description: "Repos that showed file changes to their working directory following script execution"},
	{Event: WorktreeStatusClean, Description: "Repos that showed NO file changes to their working directory following script execution"},
	{Event: CommitChangesFailed, Description: "Repos whose file changes failed to be comitted for some reason"},
	{Event: PushBranchFailed, Description: "Repos whose tool-specific branch containing changes failed to push to remote origin"},
	{Event: PushBranchSkipped, Description: "Repos whose local branch was not pushed because the --dry-run flag was set"},
	{Event: RepoNotExists, Description: "Repos that were passed via file but don't exist (404'd) via Github API"},
	{Event: PullRequestOpenErr, Description: "Repos against which pull requests failed to be opened"},
}

// RunStats will be a stats-tracker class that keeps score of which repos were touched, which were considered for update, which had branches made, PRs made, which were missing workflows or contexts, or had out of date workflows syntax values, etc
type RunStats struct {
	repos             map[Event][]*github.Repository
	pulls             map[string]string
	fileProvidedRepos []*AllowedRepo
}

// NewStatsTracker initializes a tracker struct that is capable of keeping tabs on which repos were handled and how
func NewStatsTracker() *RunStats {
	var fpr []*AllowedRepo

	t := &RunStats{
		repos:             make(map[Event][]*github.Repository),
		pulls:             make(map[string]string),
		fileProvidedRepos: fpr,
	}
	return t
}

// SetFileProvidedRepos sets the number of repos that were provided via file by the user on startup (as opposed to looked up via Github API via the --github-org flag)
func (r *RunStats) SetFileProvidedRepos(fileProvidedRepos []*AllowedRepo) {
	for _, ar := range fileProvidedRepos {
		r.fileProvidedRepos = append(r.fileProvidedRepos, ar)
	}
}

// GetMultiple returns the slice of pointers to Github repositories filed under the provided event's key
func (r *RunStats) GetMultiple(event Event) []*github.Repository {
	return r.repos[event]
}

// TrackSingle accepts an Event to associate with the supplied repo so that a final report can be generated at the end of each run
func (r *RunStats) TrackSingle(event Event, repo *github.Repository) {
	r.repos[event] = TrackEventIfMissing(r.repos[event], repo)
}

// TrackEventIfMissing prevents the addition of duplicates to the tracking slices. Repos may end up with file changes
// for example, from multiple script runs, so we don't need the same repo repeated multiple times in the final report
func TrackEventIfMissing(slice []*github.Repository, repo *github.Repository) []*github.Repository {
	for _, existingRepo := range slice {
		if existingRepo.GetName() == repo.GetName() {
			// We've already tracked this repo under this event, return the existing slice to avoid adding
			// it a second time
			return slice
		}
	}
	return append(slice, repo)
}

func (r *RunStats) TrackPullRequest(repoName, prURL string) {
	r.pulls[repoName] = prURL
}

// TrackMultiple accepts an Event and a slice of pointers to Github repos that will all be associated with that event
func (r *RunStats) TrackMultiple(event Event, repos []*github.Repository) {
	for _, repo := range repos {
		r.TrackSingle(event, repo)
	}
}

// PrintReport renders to STDOUT a summary of each repo that was considered by this tool and what happened to it during processing
func (r *RunStats) PrintReport() {
	printRepoReport(allEvents, r)
}
