#!/bin/bash

set -e

working_dir="${1:-$(pwd)}"

terraform_version=$(terraform version)
terraform_version_regex="Terraform v0\.13\..+"

if [[ "$terraform_version" =~ $terraform_version_regex ]]; then
  echo "Detected Terraform version 0.13.x"
else
  echo "ERROR: Expected Terraform version 0.13.x to be installed but found '$terraform_version'."
  exit 1
fi

echo "Finding all folders with Terraform code in '$working_dir'"

# Find all folders with Terraform files in them: https://unix.stackexchange.com/a/111952/215969
# Skip hidden files and folders (e.g., .terraform): https://askubuntu.com/a/318211
terraform_folders_str=$(find "$working_dir" -not -path '*/\.*' -type f -name '*.tf' -exec dirname {} \; | sort -u)

# Split a newline-separated string into an array: https://stackoverflow.com/a/24426608/483528
IFS=$'\n' read -d '' -ra terraform_folders_arr < <(printf '%s\0' "$terraform_folders_str")

for folder in "${terraform_folders_arr[@]}"; do
  echo "Running terraform 0.13upgrade in '$folder'"
  terraform 0.13upgrade -yes "$folder"
done