#!/usr/bin/env bash

set -eo pipefail

readonly SRC_REMOTE_NAME="origin"
readonly DST_REMOTE_NAME="internal"

readonly INTERNAL_REF_SUFFIX="internal"

function print_usage {
  echo
  echo "Usage: fork-repo.sh [ARGUMENTS]"
  echo
  echo "This script can be used to fork (i.e., copy) a Gruntwork repo into your own Git repo. The script will (a) clone the source repo into a temp folder, (b) check out each tag of the form vX.Y.Z, (c) replace cross-references to other Gruntwork repos with cross-references to your own Git repos, and (d) push the changes to a branch called v.X.Y.Z-internal in your repo. The script will also update cross-references for the master branch and push that to your repo. That way, your internal repo will have the latest code and releases and if you run this script to fork all Gruntwork repos, all cross-references should work too."
  echo
  echo "Required arguments:"
  echo
  echo -e "  --src\t\tThe URL of the Gruntwork repo to fork. This script will git clone this repo."
  echo -e "  --dst\t\tThe URL of your repo. This script will push the forked code here."
  echo -e "  --base-https\tThe base HTTPS URL for your organization. E.g., https://github.com/your-company. This is used to replace https://github.com/gruntwork-io URLs in all cross-references."
  echo -e "  --base-git\tThe base Git URL for your organization. E.g., git@github.com:your-company or github.com/your-company. This is used to replace git@github.com:gruntwork-io URLs in all cross-references."
  echo
  echo "Optional arguments:"
  echo
  echo -e "  --dry-run\t\tIf this flag is set, perform all the changes locally, but don't push them to the --dst repo. This will leave the temp folder on disk so you can inspect what would've been pushed."
  echo -e "  --dry-run-local\tSame as --dry-run, but also skip fetching data from the destination repo or checking if branches already exist. This lets you test locally without creating a destination repo."
  echo -e "  --help\t\tShow this help text and exit."
  echo
  echo "Example:"
  echo
  echo "  fork-repo.sh --src git@github.com:gruntwork-io/module-ci --dst git@github.com:your-company/module-ci --base-https https://github.com/your-company --base-git git@github.com/your-company"
}

# Log the given message at the given level. All logs are written to stderr with a timestamp.
function log {
  local -r level="$1"
  shift
  local -r message=("$@")
  local -r timestamp=$(date +"%Y-%m-%d %H:%M:%S")
  local -r script_name="$(basename "$0")"
  >&2 echo -e "${timestamp} [${level}] [$script_name] ${message[@]}"
}

# Log the given message at INFO level. All logs are written to stderr with a timestamp.
function log_info {
  log "INFO" "$@"
}

# Log the given message at ERROR level. All logs are written to stderr with a timestamp.
function log_error {
  log "ERROR" "$@"
}

# If the given value is empty, print usage instructions and exit with an error.
function assert_not_empty {
  local -r arg_name="$1"
  local -r arg_value="$2"

  if [[ -z "$arg_value" ]]; then
    log_error "The value for '$arg_name' cannot be empty"
    print_usage
    exit 1
  fi
}

# Check out the specified src repo in the current working directory. Add both the src and dst repos as a Git remotes.
function clone_repo {
  local -r src_url="$1"
  local -r dst_url="$2"

  git init 1>&2
  git remote add "$SRC_REMOTE_NAME" "$src_url" 1>&2
  git remote add "$DST_REMOTE_NAME" "$dst_url" 1>&2
  git fetch "$SRC_REMOTE_NAME" 1>&2
  git pull "$SRC_REMOTE_NAME" master 1>&2
}

# Returns 0 if the given item (needle) is in the given array (haystack); returns 1 otherwise.
function array_contains {
  local -r needle="$1"
  shift
  local -ra haystack=("$@")

  local item
  for item in "${haystack[@]}"; do
    if [[ "$item" == "$needle" ]]; then
      return 0
    fi
  done

  return 1
}

# Recursively go through all files and folders in the current working directory that match the given include patterns
# and replace the given text with the given replacement. Under the hood, we use grep for finding matching files, so you
# can use regex in the text to replace and wild cards (e.g., *.tf) in the include patterns, and we use sed for
# replacement, so you can use regex in the text to replace and capture groups in the replacement.
function replace_recursively {
  local -r text_to_replace="$1"
  local -r replacement="$2"
  shift 2
  local -ar include_patterns=("$@")

  local -a grep_opts=("-rl" "--exclude-dir=.git")
  local include_pattern
  for include_pattern in "${include_patterns[@]}"; do
    grep_opts+=("--include=$include_pattern")
  done

  grep "${grep_opts[@]}" "$text_to_replace" . | xargs sed -i '' -e "s|$text_to_replace|$replacement|g"
}

# Find all URLs pointing to Gruntwork repos and update them to point to the given URLs. Update all ref parameters to
# point to internal branches.
function update_cross_links {
  local -r base_https="$1"
  local -r base_git="$2"

  # Replace all Gruntwork Git/SSH URLs. Note that we have some repos in the HashiCorp GitHub org, so we have to replace
  # those too.
  replace_recursively "git@github.com:gruntwork-io" "$base_git" "*.*"
  replace_recursively "git@github.com:hashicorp" "$base_git" "*.*"

  # Replace all Gruntwork Git/HTTPS URLs. Note that we have some repos in the HashiCorp GitHub org, so we have to replace
  # those too. Also, note that sed doesn't support optional groups (the '?' in regex), so to handle URLs with and
  # without www, we have to essentially run the search/replace twice.
  replace_recursively "https://github.com/gruntwork-io" "$base_https" "*.*"
  replace_recursively "https://www.github.com/gruntwork-io" "$base_https" "*.*"
  replace_recursively "https://github.com/hashicorp" "$base_https" "*.*"
  replace_recursively "https://www.github.com/hashicorp" "$base_https" "*.*"

  # Replace all Terraform/Terragrunt ref parameters with internal refs
  # Example: source = "git@github.com:/gruntwork-io/module-security?ref=v0.3.4
  replace_recursively "\(source[[:space:]]*=[[:space:]]*\".*\)?ref=\(.*\)\"" "\1?ref=\2-$INTERNAL_REF_SUFFIX\"" "*.tf" "*.hcl"

  # Replace version ("tag") numbers in gruntwork-install calls in Packer templates, bash scripts, and Dockerfiles.
  # Example: gruntwork-install --module-name 'xxx' --repo 'yyy' --tag 'zzz'
  replace_recursively "\(gruntwork-install.*--tag.*v[[:digit:]]*\.[[:digit:]]*\.[[:digit:]]*\)" "\1-$INTERNAL_REF_SUFFIX" "*.json" "*.sh" "Dockerfile"

  # Replace gruntwork-installer version
  # Example: curl -Ls https://raw.githubusercontent.com/gruntwork-io/gruntwork-installer/master/bootstrap-gruntwork-installer.sh | bash /dev/stdin --version 'v0.0.21'
  replace_recursively "\(curl.*bootstrap-gruntwork-installer.sh.*v[[:digit:]]*\.[[:digit:]]*\.[[:digit:]]*\)" "\1-$INTERNAL_REF_SUFFIX" "*.json" "*.sh" "Dockerfile"

  # Replace version ("tag") numbers in bash scripts we call from Packer templates to install Gruntwork dependencies.
  # These scripts pass a version number to a function and that function calls gruntwork-install, so here, we look
  # for functions of this sort, and try to update the version numbers in them.
  # Example: install_security_packages "v0.4.5"
  replace_recursively "\(install_.*v[[:digit:]]*\.[[:digit:]]*\.[[:digit:]]*\)" "\1-$INTERNAL_REF_SUFFIX" "*.sh"
}

# Returns 0 if there are changes (diffs) in the current repo and 1 if there are no changes.
# https://stackoverflow.com/a/3899339/483528
function changes_present {
  if git diff-index --quiet HEAD; then
    return 1
  else
    return 0
  fi
}

# If there are changes on the current branch, commit them. If not, do nothing.
function commit_changes_if_necessary {
  local -r dst="$1"
  local -r branch_name="$2"

  if ! changes_present; then
    log_info "No cross-references were updated for branch '$branch_name'"
    return
  fi

  log_info "Committing updated cross-references to branch '$branch_name'"
  git add . 1>&2
  git commit -m "fork-repo.sh: automatically update cross-references to point to $dst." 1>&2
}

# Checkout the specified ref, update cross links within it, commit changes to a new internal branch, and print the name
# of the branch to stdout. Note that if the internal branch for the specified ref already exists in the destination
# repo, then this function will skip processing and print nothing to stdout.
function process_ref {
  local -r full_ref="$1"
  local -r dst="$2"
  local -r base_https="$3"
  local -r base_git="$4"
  shift 4
  local -ar dst_refs=("$@")

  local -r short_ref=$(basename "$full_ref")
  local internal_ref="$short_ref-$INTERNAL_REF_SUFFIX"

  if [[ "$short_ref" == "master" ]]; then
    log_info "Checking out master branch"
    internal_ref="master"
    git checkout master 1>&2
  else
    log_info "Creating branch '$internal_ref' from '$full_ref'."
    git checkout -B "$internal_ref" "$full_ref" 1>&2
  fi

  # We ONLY push code once to any given branch. Our cross-linked changes only go into branches in the internal repo,
  # and trying to maintain mutable branches there will likely lead to a nightmare off merge conflicts, so for now, we
  # avoid it entirely, and assume that the (immutable) tags are the only thing a customer needs, and once those are
  # pushed, there's no need to update them again. We also push master, just so the repo has a reasonable default branch,
  # but we only push it the very first time around, and do not update it after.
  if array_contains "refs/heads/$internal_ref" "${dst_refs[@]}"; then
    log_info "Branch '$internal_ref' already exists in '$dst', so will not process it again."
    return
  fi

  log_info "Updating cross links in '$full_ref' and committing changes to branch '$internal_ref'"
  update_cross_links "$base_https" "$base_git"
  commit_changes_if_necessary "$dst" "$internal_ref"

  echo -n "$internal_ref"
}

# If the --dry-run flag is set, do nothing. Otherwise, push changes to the dst repo and delete the temp check out dir.
function push_changes {
  local -r repo_path="$1"
  local -r dst="$2"
  local -r dry_run="$3"
  local -r dry_run_local="$4"
  shift 4
  local -ar refs_to_push=("$@")

  log_info "The following ${#refs_to_push[@]} branches were updated: ${refs_to_push[@]}"

  if [[ "$dry_run" == "true" || "$dry_run_local" == "true" ]]; then
    log_info "The --dry-run or --dry-run-local flag is set, so will not 'git push' changes and will skip deleting the temp checkout dir so you can inspect the results: $repo_path"
    return
  elif [[ -z "${refs_to_push[@]}" ]]; then
    log_info "No branches were updated, so nothing to push!"
  else
    log_info "Pushing changes to '$dst'"
    git push "$DST_REMOTE_NAME" "${refs_to_push[@]}" 1>&2
  fi

  log_info "Cleaning up tmp checkout dir $repo_path"
  rm -rf "$repo_path"
}

function run {
  local src
  local dst
  local base_https
  local base_git
  local dry_run="false"
  local dry_run_local="false"

  if [[ "$#" == 0 ]]; then
    print_usage
    exit
  fi

  while [[ $# > 0 ]]; do
    local key="$1"

    case "$key" in
      --src)
        src="$2"
        shift
        ;;
      --dst)
        dst="$2"
        shift
        ;;
      --base-https)
        base_https="$2"
        shift
        ;;
      --base-git)
        base_git="$2"
        shift
        ;;
      --dry-run)
        dry_run="true"
        ;;
      --dry-run-local)
        dry_run_local="true"
        ;;
      --help)
        print_usage
        exit
        ;;
    esac

    shift
  done

  assert_not_empty "--src" "$src"
  assert_not_empty "--dst" "$dst"
  assert_not_empty "--base-https" "$base_https"
  assert_not_empty "--base-git" "$base_git"

  local repo_path
  repo_path=$(mktemp -d -t fork-repo)

  (cd "$repo_path" && clone_repo "$src" "$dst")

  # Get all tags in the src repo
  local -a src_refs
  src_refs=($(cd "$repo_path" && git ls-remote --tags "$SRC_REMOTE_NAME" | cut -f2))

  # Get all branches in the dest repo
  local -a dst_refs
  if [[ "$dry_run_local" == "true" ]]; then
    log_info "The --dry-run-local flag is set, so will not check the destination repo for existing branches, and will process everything from scratch."
    dst_refs=()
  else
    dst_refs=($(cd "$repo_path" && git ls-remote --heads "$DST_REMOTE_NAME" | cut -f2))
  fi

  # Add the master branch to the list of src refs, as we always want to copy the latest code for master
  src_refs=("refs/heads/master" "${src_refs[@]}")

  local -a refs_to_push=()
  local src_ref
  local dst_ref

  for src_ref in "${src_refs[@]}"; do
    dst_ref=$(cd "$repo_path" && process_ref "$src_ref" "$dst" "$base_https" "$base_git" "${dst_refs[@]}")
    if [[ ! -z "$dst_ref" ]]; then
      refs_to_push+=("$dst_ref")
    fi
  done

  (cd "$repo_path" && push_changes "$repo_path" "$dst" "$dry_run" "$dry_run_local" "${refs_to_push[@]}")
}

run "$@"
