#!/usr/bin/env bash

set -eo pipefail

# Gruntwork repos in the gruntwork-io org
readonly GRUNTWORK_REPOS=(
  # AWS repos
  "gruntkms"
  "infrastructure-live-acme"
  "infrastructure-live-google"
  "infrastructure-live-multi-account-acme"
  "infrastructure-modules-acme"
  "infrastructure-modules-google"
  "infrastructure-modules-multi-account-acme"
  "module-asg"
  "module-aws-monitoring"
  "module-cache"
  "module-ci"
  "module-data-storage"
  "module-ecs"
  "module-load-balancer"
  "module-security"
  "module-server"
  "module-vpc"
  "package-beanstalk"
  "package-elk"
  "package-kafka"
  "package-lambda"
  "package-messaging"
  "package-openvpn"
  "package-static-assets"
  "package-terraform-utilities"
  "package-zookeeper"
  "package-mongodb"
  "package-sam"
  "sample-app-backend-acme"
  "sample-app-backend-multi-account-acme"
  "sample-app-frontend-acme"
  "sample-app-frontend-multi-account-acme"
  "terraform-aws-couchbase"
  "terraform-aws-eks"
  "terraform-aws-influx"

  # GCP repos
  "terraform-google-gke"
  "terraform-google-influx"
  "terraform-google-load-balancer"
  "terraform-google-network"
  "terraform-google-security"
  "terraform-google-sql"
  "terraform-google-static-assets"

  # Kubernetes repos
  "helm-kubernetes-services"
  "kubergrunt"
  "terraform-helm-gke-exts"
  "terraform-kubernetes-helm"

  # Common tools
  "bash-commons"
  "fetch"
  "gruntwork"
  "gruntwork-cli"
  "gruntwork-installer"
  "terragrunt"
  "terratest"
)

# Gruntwork repos in the hashicorp GitHub org
readonly GRUNTWORK_HASHICORP_REPOS=(
  # AWS repos
  "terraform-aws-consul"
  "terraform-aws-nomad"
  "terraform-aws-vault"

  # GCP repos
  "terraform-google-consul"
  "terraform-google-nomad"
  "terraform-google-vault"
)

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

function print_usage {
  echo
  echo "Usage: fork-all-repos.sh [ARGUMENTS]"
  echo
  echo "This script can be used to fork (i.e., copy) all Gruntwork repos into your own Git repos. This script loops over each Gruntwork repo github.com/gruntwork-io/<NAME>, runs fork-repo.sh on it to update all cross-references and tags, and pushes the changes to <YOUR_GIT_URL>/<NAME><YOUR_SUFFIX>. See the fork-repo.sh script or how cross-references and tags are updated."
  echo
  echo "Required arguments:"
  echo
  echo -e "  --base-https\tThe base HTTPS URL for your organization. E.g., https://github.com/your-company. This is used to replace https://github.com/gruntwork-io URLs in all cross-references."
  echo -e "  --base-git\tThe base Git URL for your organization. E.g., git@github.com:your-company or github.com/your-company. This is used to replace git@github.com:gruntwork-io URLs in all cross-references."
  echo
  echo "Optional arguments:"
  echo
  echo -e "  --suffix\t\tIf specified, this suffix will be appended to every repo name. That is, each Grunwork repo foo will be pushed to an internal repo of yours called foo<SUFFIX>. Default: (empty string)."
  echo -e "  --dry-run\t\tIf this flag is set, perform all the changes locally, but don't git push them. This will leave the temp folders on disk so you can inspect what would've been pushed."
  echo -e "  --dry-run-local\tSame as --dry-run, but also skip fetching data from the destination repo or checking if branches already exist. This lets you test locally without creating a destination repo."
  echo -e "  --help\t\tShow this help text and exit."
  echo
  echo "Example:"
  echo
  echo "  fork-all-repos.sh --base-https https://github.com/your-company --base-git git@github.com/your-company"
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

function process_repo {
  local -r src_repo="$1"
  local -r src_url="$2"
  local -r base_https="$3"
  local -r base_git="$4"
  local -r suffix="$5"
  local -r dry_run="$6"
  local -r dry_run_local="$7"

  local -r dst_repo="${src_repo}${suffix}"
  local -r dst_url="$base_git/$dst_repo.git"

  local -a args=(
    --src "$src_url"
    --dst "$dst_url"
    --base-https "$base_https"
    --base-git "$base_git"
  )

  if [[ "$dry_run" == "true" ]]; then
    args+=("--dry-run")
  fi
  if [[ "$dry_run_local" == "true" ]]; then
    args+=("--dry-run-local")
  fi

  log_info "Forking repo '$src_url' to '$dst_url'"

  "$SCRIPT_DIR/fork-repo.sh" "${args[@]}"
}

function run {
  local base_https
  local base_git
  local suffix=""
  local dry_run="false"
  local dry_run_local="false"

  if [[ "$#" == 0 ]]; then
    print_usage
    exit
  fi

  while [[ $# > 0 ]]; do
    local key="$1"

    case "$key" in
      --base-https)
        base_https="$2"
        shift
        ;;
      --base-git)
        base_git="$2"
        shift
        ;;
      --suffix)
        suffix="$2"
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

  assert_not_empty "--base-https" "$base_https"
  assert_not_empty "--base-git" "$base_git"

  local src_repo
  local dst_repo
  local src_url
  local dst_url
  local -a args=()

  for src_repo in "${GRUNTWORK_REPOS[@]}"; do
    process_repo "$src_repo" "git@github.com:gruntwork-io/$src_repo.git" "$base_https" "$base_git" "$suffix" "$dry_run" "$dry_run_local"
  done

  for src_repo in "${GRUNTWORK_HASHICORP_REPOS[@]}"; do
    process_repo "$src_repo" "git@github.com:hashicorp/$src_repo.git" "$base_https" "$base_git" "$suffix" "$dry_run" "$dry_run_local"
  done
}

run "$@"
