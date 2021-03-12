# Overview

If you have have some logic or set of changes you want to execute across a number of repos:

1. write your logic in a single script (bash, ruby, python, etc are supported!)
1. write the script so that it operates on a single repo
1. use this tool to execute your script against all the repos you select
1. optionally pass your desired commit messages, pull request titles and pull request bodies
1. the tool will execute your scripts in the order you provided them, and handle all resulting git operations for you, opening PRs with the changes
1. a final report of what happened with each repo and the links to the resulting PRs will be output at the end of your run

You can think of this tool as [xargs](https://en.wikipedia.org/wiki/Xargs) for git!

# How it works

The multi repo script runner is a CLI that allows you to quickly make mass updates to multiple github repositories by:
* Allowing you to write arbitrary scripts (bash, ruby, python, etc)
* Allowing you to select multiple Github repos to target by supplying either a). a Github organization name or b). a flat file containing repo names
* Cloning each of your selected repos to your /tmp/ directory and creating a new branch
* Running the bash scripts you specify via the `--scripts="add-license.sh,my-other-script-too.sh, /tmp/my-ruby-script.rb, ./scripts/my-relative-python-script.py"` flag (relative and absolute paths are supported!)
* Commiting any file additions, deletions or untracked files that result, using a configurable commit message
* Pushing the branch containing your changes to the remote origin
* Opening a pull request using configurable PR title and PR description

# Example tasks this tool is well-suited for

* Add an LICENSE file to all of your GitHub repos, interpolating the correct year and company name into the file
* For every existing LICENSE file across all your repos, update their copyright date to the current year
* Update the CI build configuration in all of your repos by modifying the `.circleci/config.yml` file in each repo using a tool such as `yq`
* Perform modifications on any JSON files via tools such as `jq`
* Run `sed` commands to update or replace information across README files
* Add new files to repos
* Delete specific files, when present, from repos
* Really - anything you feel like implementing in a bash script that can be done to a single repository!

# Getting started

## Export a valid Github token

See [Github personal access token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token) for information on how to generate one. Note that you must do this while logged into a Github account that is a member of the Gruntwork organization so that it has access to our private repos.

```
export GITHUB_OAUTH_TOKEN
```

## Build the binary

```
go build
./multi-repo-script-runner -h
```
Running the help command will output the following:

```
Multi repo script runner executes arbitrary bash scripts against any repos you select, handling all git operations that result and opening configurable pull requests

Usage:
  multi-repo-script-runner [flags]
  multi-repo-script-runner [command]

Available Commands:
  help        Help about any command
  version     Print the multi-repo-script-runner's version number

Flags:
  -a, --allowed-repos-filepath string     The path to the file containing repos this tool is allowed to operate on, each repo in format: gruntwork-io/terraform-aws-eks, one repo per line
  -b, --branch-name string                The name of the branch you want created to hold your changes (default "multi-repo-script-runner")
  -m, --commit-message string             The commit message to use for any programmatic commits made by this tool (default "Tis I, the multi-repo-script-runner!")
  -d, --dry-run                           When dry-run is set to true, only proposed YAML updates will be output, but not changes in Github will be made (no branches will be created, no files updated, no PRs opened)
  -o, --github-org string                 The Github organization whose repos should be operated on
  -h, --help                              help for multi-repo-script-runner
  -e, --pull-request-description string   The description to add to the pull requests that will be opened by this run (default "This pull request was opened programmatically by the multi repo script runner CLI.")
  -t, --pull-request-title string         The title to add to the pull requests that will be opened by this run (default "Multi Repo Updater Programmatic PR")
  -s, --scripts strings                   The scripts to run against the selected repos. These scripts must exist in the ./scripts directory and be executable.
```
## Run the tool without building the binary

Alternatively, you can run the tool directly without building the binary, like so:

`./go run main.go --commit-message "Add MIT License" --pull-request-title "Add MIT License" --pull-request-description "These changes add an MIT license file to repo, including the correct year and Gruntwork, Inc as the full name" --scripts="./scripts/add-license.sh" --allowed-repos-filepath data/zack-test-repos.txt`

This is especially helpful if you are developing against the tool and want to quickly verify your changes.

## Selecting scripts to run, and a note on paths
Use the `--scripts=one-script.sh,two-script.sh,red-script.sh,blue-script.rb` flag (shorthand is `-s="<script1>,<script2>"`) to select the scripts to run against each of the selected repos. Note that, because the tool supports bash scripts, ruby scripts, python scripts, etc, you must include the full filename for any given script, including its file extension.

Scripts may be placed anywhere on your system, and the tool will accept relative and absolute paths to scripts, and they can be intermixed in a single command. For example, you may choose to version some scripts in the `./scripts` directory of this tool so that everyone has access to them, in which case you can pass `-s="./scripts/versioned-script.rb, /tmp/some-other-script.sh, /home/zachary/Code/project/script.py"` all in the same run.

## Selecting repos to run your scripts against
There are two options for selecting repos to run your scripts against:
1. Pass the `--github-org` option followed by the name of the Github org to look up repos for. e.g., `--github-org gruntwork-io`. This will page through ALL the repos in the selected organization, running your selected scripts on EACH of them
1. If you need more control over which repos to execute your scripts against, you can define a flat file of the exact repos to select and pass the `--allowed-repos-filepath` flag (`-a`) like so: `-a data/zack-test-repos.txt`
	1. The flatfile must be formatted with one repo per line in the following format `gruntwork-io/cloud-nuke`
	1. Trailing commas are options, and preceding or trailing space is irrelevant, as are single and double quotes

## Handling prerequisites and third party binaries

It is currently assumed that bash script authors will be responsible for checking for prequisites within their own scripts. If you are adding a new bash script to accomplish some new task across repos, consider using the [Gruntwork bash-commons assert_is_installed pattern](https://github.com/gruntwork-io/bash-commons/blob/3cb3c7160fb72b7411af184300bf077caede37e4/modules/bash-commons/src/assert.sh#L15) to ensure the operator has any required binaries installed.

That said, this CLI does have a method of requiring binaries to be installed and in the system PATH at runtime. You may also patch [that system to add a new Dependency, following the established pattern.](https://github.com/gruntwork-io/prototypes/pull/96/files#diff-4ff8ecb1e2d8ab5644a567ac1553dcbc41302b96949d488a189fa0e927276e97R71-R83)

## Examples

### Add a new license file to every repo defined in a flatfile

Note that the actual logic is implemented in `./scripts/add-license.sh`

`./multi-repo-script-runner --commit-message "Add MIT License" --pull-request-title "Add MIT License" --pull-request-description "These changes add an MIT license file to repo, including the correct year and Gruntwork, Inc as the full name" --scripts="add-license" --allowed-repos-filepath data/zack-test-repos.txt`

# Background

This tool is an offshoot of the context upgrader CLI that was implemented in [IAC-1616](https://gruntwork.atlassian.net/browse/IAC-1616). However, this iteration is far more flexible and powerful because it allows you to implement any logic you want via bash scripts, and then handle the mass updates for you by executing your scripts and handling any resultant git tasks for you. We've also discussed this tool as a form of [xargs for git](https://www.notion.so/gruntwork/An-xargs-for-updating-multiple-Git-repos-f3abbf4b1c2b4dd597cd122c50c10c82#2dd15aa30caf48388d47a120b3720757).
