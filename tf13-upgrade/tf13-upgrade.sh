#!/bin/bash

set -e

working_dir="${1:-$(pwd)}"

if ! command -v "tfenv" > /dev/null; then
  echo "[ERROR] This script requires that tfenv (https://github.com/tfutils/tfenv) is installed! Run: brew install tfenv."
  exit 1
fi

terraform_version_constraint="latest:^0.13"

echo "Using tfenv to switch Terraform version to '$terraform_version_constraint'"
set +e
output=$(tfenv use "$terraform_version_constraint" 2>&1 | tee /dev/tty)
exit_code="$?"
set -e

if [[ "$exit_code" -ne 0 ]]; then
  if [[ "$output" == *"No installed versions of terraform matched"* ]]; then
    echo "Looks like no Terraform version matching '$terraform_version_constraint' is installed. Using tfenv to install it."
    tfenv install "$terraform_version_constraint"
    echo "Using tfenv to switch Terraform version to '$terraform_version_constraint'"
    tfenv use "$terraform_version_constraint"
  else
    echo "[ERROR] Expected 'tfenv use' to exit with status code 0 but got $exit_code. Log output is above."
    exit 1
  fi
fi

echo "Finding all folders with Terraform code in '$working_dir'"

# Find all folders tracked by Git with Terraform files in them: https://stackoverflow.com/a/20247815/483528
# Use xargs with dirname to get the parent folder of each Terraform file we find.
# Use sort and uniq to remove any duplicate folders.
terraform_folders_str=$(git --git-dir "$working_dir/.git" ls-files **/*.tf *.tf | xargs -n 1 dirname | sort | uniq)

# Split a newline-separated string into an array: https://stackoverflow.com/a/24426608/483528
IFS=$'\n' read -d '' -ra terraform_folders_arr < <(printf '%s\0' "$terraform_folders_str")

required_version_regex='required_version[[:space:]]*=[[:space:]]*".*"'
required_version_replacement_first_line="This module is now only being tested with Terraform 0.13.x. However, to make upgrading easier, we are setting"
# Note that the backslashes and newlines here in the middle of the string below are intentional! That's because we use
# this variable with sed, and on MacOS, sed does not support '\n' in replacement text, but a backslash followed by a
# literal newline should work: https://superuser.com/a/307486
required_version_replacement="# $required_version_replacement_first_line\\
  # 0.12.26 as the minimum version, as that version added support for required_providers with source URLs, making it\\
  # forwards compatible with 0.13.x code.\\
  required_version = \">= 0.12.26\""
required_version_replacement_without_slashes="${required_version_replacement//\\/}"

tf13_upgrade_modules_with_errors=()
tf13_upgrade_modules_with_errors_log_output=()
missing_required_version=()

for path in "${terraform_folders_arr[@]}"; do
  folder="$working_dir/$path"
  main_file="$folder/main.tf"
  versions_file="$folder/versions.tf"

  # Try to make this script idempotent by not re-upgrading code that has already been upgraded and has our comment in
  # it indicating that we've patched the versions.tf file.
  if grep -q "$required_version_replacement_first_line" "$main_file" > /dev/null 2>&1; then
    echo "We've already upgraded the code in '$folder'. Will not upgrade again."
  else
    echo "Running terraform 0.13upgrade in '$folder'"

    # To reduce log output noise, we capture stdout/stderr from the 0.13.upgrade command. However, if it exits in an
    # error, we want to show the log output to help with debugging. Therefore, we temporarily disable -e, and check
    # for errors manually.
    set +e
    output=$(terraform 0.13upgrade -yes "$folder" 2>&1)
    exit_code="$?"
    set -e

    if [[ "$exit_code" -ne 0 ]]; then
      echo "[ERROR] Expected the terraform 0.13upgrade command to exit with code 0, but it exited with code $exit_code. Log output from the command is shown below."
      echo -e "$output"
      tf13_upgrade_modules_with_errors+=("$folder")
      tf13_upgrade_modules_with_errors_log_output+=("$output")
    fi

    if [[ -f "$versions_file" ]]; then
      # The way the 0.13upgrade command handles version upgrade is... quite poor. Instead of spending loads of time
      # trying to clean it up, we just delete the versions.tf file, as it's not necessary. See the discussion at
      # https://github.com/gruntwork-io/prototypes/pull/75#discussion_r488787369 for more context.
      echo "Deleting the versions.tf file created by the 0.13upgrade command."
      rm -f "$versions_file"
    fi

    if [[ -f "$main_file" ]] && grep -q "$required_version_regex" "$main_file" > /dev/null 2>&1; then
      # We set the required_version to 0.12.26, as that version supports required_providers with source URLs, so it's
      # forward compatible with Terraform 0.13.x. Although we'll only be testing our code with Terraform 0.13.x after
      # the upgrade, allowing 0.12.26 and above will give our users more time to do the upgrade.
      echo "Overwriting version constraint in '$main_file' to support TF 0.12.x."
      sed -i '' "s/$required_version_regex/$required_version_replacement/g" "$main_file"
    fi
  fi

  if ! grep -q "$required_version_regex" "$main_file" > /dev/null 2>&1; then
    echo "[WARN] Did not find required_version in '$main_file'."
    missing_required_version+=("$main_file")
  fi
done

echo
echo "Next steps:"
echo

if [[ -n "${tf13_upgrade_modules_with_errors[*]}" ]]; then
  echo '===== terraform 0.13upgrade errors ====='
  echo
  echo -e "There were ${#tf13_upgrade_modules_with_errors[@]} Terraform modules with errors when we ran the 0.13upgrade command. The list of paths and their log output is below. Go through each one and fix the issues!"
  echo
  for (( i=0; i<"${#tf13_upgrade_modules_with_errors[@]}"; i++ )); do
    echo "== Module $(( i+1 ))/${#tf13_upgrade_modules_with_errors[@]}: ${tf13_upgrade_modules_with_errors[i]} =="
    echo -e "\n\n${tf13_upgrade_modules_with_errors_log_output[i]}\n\n"
  done
  echo
fi

if [[ -n "${missing_required_version[*]}" ]]; then
  echo '===== required_version usage ====='
  echo
  echo -e "Did not find a terraform { ... } block with a 'required_version' param in the files below. Please add the following block to the files below:\n\nterraform {\n  $required_version_replacement_without_slashes\n}\n"
  echo
  for file in "${missing_required_version[@]}"; do
    echo "- $file"
  done
  echo
fi

echo "===== Commit ====="
echo
echo "Once all the above is done, do the following:"
echo
echo "- Check the diffs in Git"
echo "- Test the code"
echo "- Submit a PR"
echo
