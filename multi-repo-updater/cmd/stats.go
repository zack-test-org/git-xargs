package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/kataras/tablewriter"
	"github.com/landoop/tableprinter"
)

// Event is a generic tracking ocurrence that RunStats manages
type Event string

const (
	// ConfigFound denotes a repo has the Circle CI config at the expected path
	ConfigFound Event = "circle-ci-config-found"
	// ConfigNotFound denotes a repo that was missing its CirlceCI config
	ConfigNotFound Event = "circle-ci-config-not-found"
	// ContextAlreadySet denotes a repo whose jobs already have the target context set
	ContextAlreadySet Event = "circle-ci-config-contexts-already-set"
	// DryRunSet denotes a repo will not have any file changes, branches made or PRs opened because the dry-run flag was set to true
	DryRunSet Event = "dry-run-set-no-changes-made"
	// TargetBranchNotFound denotes the special branch used by this tool to make changes on was not found on lookup, suggesting it should be created
	TargetBranchNotFound Event = "target-branch-not-found"
	// TargetBranchAlreadyExists denotes the special branch used by this tool was already found (so it was likely already created by a previous run)
	TargetBranchAlreadyExists Event = "target-branch-already-exists"
	// TargetBranchLookupErr denotes an issue performing the lookup via Github API for the target branch - an API call failure
	TargetBranchLookupErr Event = "target-branch-lookup-err"

	// YamlNotUpdated denotes that this tool determined it did not need to, or could not, programmatically modify the given repo's config file
	YamlNotUpdated Event = "yaml-not-updated"
)

// AnnotatedEvent is used in printing the final report. It contains the info to print a section's table - both it's Event for looking up the tagged repos, and the human-legible description for printing above the table
type AnnotatedEvent struct {
	Event       Event
	Description string
}

var allEvents = []AnnotatedEvent{
	{Event: ConfigFound, Description: "Repos with Circle CI config files"},
	{Event: ConfigNotFound, Description: "Repos that did not have Circle CI config files"},
	{Event: ContextAlreadySet, Description: "Repos that already had the correct context set"},
	{Event: DryRunSet, Description: "Repos that were not modified in any way because this was a dry-run"},
	{Event: TargetBranchNotFound, Description: "Repos whose target branch was not found"},
	{Event: TargetBranchAlreadyExists, Description: "Repos whose target branch already existed"},
	{Event: TargetBranchLookupErr, Description: "Repos whose target branches could not be looked up due to an API error"},
	{Event: YamlNotUpdated, Description: "Repos whose config files were unmodified by this tool"},
}

// RunStats will be a stats-tracker class that keeps score of which repos were touched, which were considered for update, which had branches made, PRs made, which were missing workflows or contexts, or had out of date workflows syntax values, etc
type RunStats struct {
	repos             map[Event][]*github.Repository
	fileProvidedRepos []*AllowedRepo
}

// NewStatsTracker initializes a tracker struct that is capable of keeping tabs on which repos were handled and how
func NewStatsTracker() *RunStats {
	var fpr []*AllowedRepo

	t := &RunStats{
		repos:             make(map[Event][]*github.Repository),
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

// TrackSingle accepts an Event to associate with the supplied repo so that a final report can be generated at the end of each run
func (r *RunStats) TrackSingle(event Event, repo *github.Repository) {
	r.repos[event] = append(r.repos[event], repo)
}

// TrackMultiple accepts an Event and a slice of pointers to Github repos that will all be associated with that event
func (r *RunStats) TrackMultiple(event Event, repos []*github.Repository) {
	for _, repo := range repos {
		r.TrackSingle(event, repo)
	}
}

func configurePrinterStyling(printer *tableprinter.Printer) {
	printer.BorderTop, printer.BorderBottom, printer.BorderLeft, printer.BorderRight = true, true, true, true
	printer.CenterSeparator = "│"
	printer.ColumnSeparator = "│"
	printer.RowSeparator = "─"
	printer.HeaderBgColor = tablewriter.BgBlackColor
	printer.HeaderFgColor = tablewriter.FgGreenColor
}

// PrintReport renders to STDOUT a summary of each repo that was considered by this tool and what happened to it during processing
func (r *RunStats) PrintReport() {
	fmt.Print("\n\n")
	fmt.Println("*****************************************************")
	fmt.Printf("RUN SUMMARY @ %v\n", time.Now().UTC())
	fmt.Println("*****************************************************")

	// If there were any allowed repos provided via file, print out the list of them
	fileProvidedReposPrinter := tableprinter.New(os.Stdout)
	configurePrinterStyling(fileProvidedReposPrinter)

	fmt.Print("\n\n")
	fmt.Println("REPOS SUPPLIED VIA --allowed-repos-filepath FLAG")
	fileProvidedReposPrinter.Print(r.fileProvidedRepos)

	// For each event type, print a summary of the repos in that category
	for _, ae := range allEvents {

		var reducedRepos []ReducedRepo

		printer := tableprinter.New(os.Stdout)
		configurePrinterStyling(printer)

		for _, repo := range r.repos[ae.Event] {
			rr := ReducedRepo{
				Name: repo.GetName(),
				URL:  repo.GetHTMLURL(),
			}
			reducedRepos = append(reducedRepos, rr)
		}

		if len(reducedRepos) > 0 {
			fmt.Println()
			fmt.Println(strings.ToUpper(ae.Description))
			printer.Print(reducedRepos)
			fmt.Println()
		}
	}
}
