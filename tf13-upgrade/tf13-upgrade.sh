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

required_version_regex='required_version = ">= 0.13"'
required_version_replacement_first_line="This module is now only being tested with Terraform 0.13.x. However, to make upgrading easier, we are setting"
# Note that the backslashes and newlines here in the middle of the string below are intentional! That's because we use
# this variable with sed, and on MacOS, sed does not support '\n' in replacement text, but a backslash followed by a
# literal newline should work: https://superuser.com/a/307486
required_version_replacement="\\
  # $required_version_replacement_first_line\\
  # 0.12.20 as the minimum version, as that version added support for required_providers, making it forwards compatible\\
  # with 0.13.x code.\\
  required_version = \">= 0.12.20\""

for folder in "${terraform_folders_arr[@]}"; do
  versions_file="$folder/versions.tf"

  # Try to make this script idempotent by not re-upgrading code that has already been upgraded and has our comment in
  # it indicating that we've patched the versions.tf file.
  if grep -q "$required_version_replacement_first_line" "$versions_file" > /dev/null 2>&1; then
    echo "We've already upgraded the code in '$folder'. Will not upgrade again."
  else
    echo "Running terraform 0.13upgrade in '$folder'"
    terraform 0.13upgrade -yes "$folder"

    # The 0.13upgrade command should have created this file
    if [[ ! -f "$versions_file" ]]; then
      echo "ERROR: Expected file '$versions_file' does not exist!"
      exit 1
    fi

    echo "Overwriting version constraint in '$versions_file' to indicate the code is still compatible with TF 0.12"
    sed -i '' "s/$required_version_regex/$required_version_replacement/g" "$versions_file"
  fi
done