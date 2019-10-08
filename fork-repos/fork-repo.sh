#!/usr/bin/env bash

set -e
set -o pipefail

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
  echo -e "  --base-https\tThe base HTTPS URL for your organization. E.g., https://github.com/your-company. This is used to replace https://github.com/gruntwork-io URLs in all cross-references. "
  echo -e "  --base-git\tThe base Git URL for your organization. E.g., git@github.com:your-company. This is used to replace git@github.com:gruntwork-io URLs in all cross-references. "
  echo
  echo "Optional arguments:"
  echo
  echo -e "  --dry-run\tIf this flag is set, perform all the changes locally, but don't push them to the --dst repo. This will leave the temp folder on disk so you can inspect what would've been pushed."
  echo -e "  --help\tShow this help text and exit."
  echo
  echo "Example:"
  echo
  echo "  fork-repo.sh --src git@github.com:gruntwork-io/module-ci --dst git@github.com:your-company/module-ci --base-https https://github.com/your-company --base-git git@github.com/your-company"
}

# Log the given message at the given level. All logs are written to stderr with a timestamp.
function log {
  local -r level="$1"
  local -r message="$2"
  local -r timestamp=$(date +"%Y-%m-%d %H:%M:%S")
  local -r script_name="$(basename "$0")"
  >&2 echo -e "${timestamp} [${level}] [$script_name] ${message}"
}

# Log the given message at INFO level. All logs are written to stderr with a timestamp.
function log_info {
  local -r message="$1"
  log "INFO" "$message"
}

# Log the given message at ERROR level. All logs are written to stderr with a timestamp.
function log_error {
  local -r message="$1"
  log "ERROR" "$message"
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

# Find all URLs pointing to Gruntwork repos and update them to point to the given URLs. Update all ref parameters to
# point to internal branches.
function update_cross_links {
  local -r base_https="$1"
  local -r base_git="$2"

  # Replace Git URLs everywhere
  find . -type f -print0 | xargs -0 sed -i '' -e "s|git@github.com:gruntwork-io|$base_git|g"

  # Replace HTTPS URLs everywhere
  find . -type f -print0 | xargs -0 sed -i '' -e "s|https://github.com/gruntwork-io|$base_https|g"
  find . -type f -print0 | xargs -0 sed -i '' -e "s|https://www.github.com/gruntwork-io|$base_https|g"

  # Replace ref parameters in Terraform/Terragrunt source URLs
  find . -type f -name '*.tf' -name '*.hcl' -print0 | xargs -0 sed -i '' -e "s|\(source[[:space:]]*=[[:space:]]*\".*\)?ref=\(.*\)\"|\1?ref=\2-$INTERNAL_REF_SUFFIX\"|g"

  # TODO: Tags in Packer templates. Ideally, we'd look for gruntwork-install --tag "xxx" and replace the "xxx", but
  # many of our templates call gruntwork-install in Bash scripts, and pass the tag using a variable, so it's not
  # obvious how to replace it.
}

# Returns 0 if there are changes (diffs) in the current repo and 1 if there are no changes.
# https://stackoverflow.com/a/3899339/483528
function changes_present {
  if git diff-index --quiet HEAD; then
    return 0
  else
    return 1
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

  log_info "Updated cross-references in the following files:"
  git diff-index --name-only HEAD 1>&2

  log_info "Committing these changes to branch '$branch_name'"
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
    internal_ref="master"
    git checkout master 1>&2
  elif array_contains "$internal_ref" "$dst_refs"; then
    log_info "Branch '$internal_ref' already exists i '$dst', so will not process it again."
    return
  else
    git checkout -B "$internal_ref" "$full_ref" 1>&2
  fi

  log_info "Updating cross links in branch '$short_ref' and committing changes to branch '$internal_ref' in '$dst'"
  update_cross_links "$base_https" "$base_git"
  commit_changes_if_necessary "$dst" "$internal_ref"

  echo -n "$internal_ref"
}

# If the --dry-run flag is set, do nothing. Otherwise, push changes to the dst repo and delete the temp check out dir.
function push_changes {
  local -r repo_path="$1"
  local -r dst="$2"
  local -r dry_run="$3"
  shift 3
  local -ar refs_to_push=("$@")

  if [[ "$dry_run" == "true" ]]; then
    log_info "The --dry-run flag is set, so will not 'git push' changes. You can inspect the changes in the folder: $repo_path"
    log_info "The following branches were updated and would've been pushed if the --dry-run flag had not been set: ${refs_to_push[@]}"
    return
  elif [[ -z "${refs_to_push[@]}" ]]; then
    log_info "No refs were updated, so nothing to push!"
    return
  fi

  log_info "Pushing changes to the following branches in '$dst': ${refs_to_push[@]}"
  git push "$DST_REMOTE_NAME" "${refs_to_push[@]}" 1>&2

  log_info "Cleaning up tmp checkout dir $repo_path"
  rm -rf "$repo_path"
}

function run {
  local src
  local dst
  local base_https
  local base_git
  local dry_run="false"

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
  local src_refs
  src_refs=($(cd "$repo_path" && git ls-remote --tags "$SRC_REMOTE_NAME" | cut -f2))

  # Get all branches in the dest repo
  local dst_refs
  dst_refs=($(cd "$repo_path" && git ls-remote --heads "$DST_REMOTE_NAME" | cut -f2))

  # Add the master branch to the list of src refs, as we always want to copy the latest code for master
  src_refs=("refs/heads/master" "${src_refs[@]}")

  local refs_to_push=()
  local src_ref
  local dst_ref

  # TODO: remove count!!!
  local count=0
  for src_ref in "${src_refs[@]}"; do
    dst_ref=$(cd "$repo_path" && process_ref "$src_ref" "$dst" "$base_https" "$base_git" "${dst_refs[@]}")
    if [[ ! -z "$dst_ref" ]]; then
      refs_to_push+=("$dst_ref")
    fi

    # TODO: remove count!!!
    count=$((count+1))
    if [[ "$count" -gt 5 ]]; then
      log_info "TEMPORARILY STOPPING. REMOVE THIS SHIT BEFORE COMMITTING."
      break
    fi
  done

  (cd "$repo_path" && push_changes "$repo_path" "$dst" "$dry_run" "$refs_to_push")
}

run "$@"