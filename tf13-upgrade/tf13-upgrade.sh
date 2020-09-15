#!/bin/bash

set -e

working_dir="${1:-$(pwd)}"

terraform_version=$(terraform version)
terraform_version_regex="Terraform v0\.13\..+"

if [[ "$terraform_version" =~ $terraform_version_regex ]]; then
  echo "Detected Terraform version 0.13.x"
else
  echo "[ERROR] Expected Terraform version 0.13.x to be installed but found '$terraform_version'."
  exit 1
fi

echo "Finding all folders with Terraform code in '$working_dir'"

# Find all folders tracked by Git with Terraform files in them: https://stackoverflow.com/a/20247815/483528
terraform_folders_str=$(git --git-dir "$working_dir/.git" ls-files *.tf | sed s,/[^/]*$,, | uniq)

# Split a newline-separated string into an array: https://stackoverflow.com/a/24426608/483528
IFS=$'\n' read -d '' -ra terraform_folders_arr < <(printf '%s\0' "$terraform_folders_str")

required_version_regex='required_version = ">= 0.13"'
required_version_replacement_first_line="This module is now only being tested with Terraform 0.13.x. However, to make upgrading easier, we are setting"
# Note that the backslashes and newlines here in the middle of the string below are intentional! That's because we use
# this variable with sed, and on MacOS, sed does not support '\n' in replacement text, but a backslash followed by a
# literal newline should work: https://superuser.com/a/307486
required_version_replacement="\\
  # $required_version_replacement_first_line\\
  # 0.12.26 as the minimum version, as that version added support for required_providers with source URLs, making it\\
  # forwards compatible with 0.13.x code.\\
  required_version = \">= 0.12.26\""

legacy_required_version_uses=()
destroy_provisioner_uses=()

for path in "${terraform_folders_arr[@]}"; do
  folder="$working_dir/$path"
  main_file="$folder/main.tf"
  versions_file="$folder/versions.tf"

  # Try to make this script idempotent by not re-upgrading code that has already been upgraded and has our comment in
  # it indicating that we've patched the versions.tf file.
  if grep -q "$required_version_replacement_first_line" "$versions_file" > /dev/null 2>&1; then
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
      exit "$exit_code"
    fi

    if [[ ! -f "$versions_file" ]]; then
      # This usually only happens in super simple Terraform examples that use no providers.
      echo "The 0.13upgrade command did not create a versions.tf file. Creating one at '$versions_file'."
      echo -e "terraform {\n  $required_version_regex\n}\n" > "$versions_file"
    fi

    echo "Overwriting version constraint in '$versions_file' to support TF 0.12.x."
    sed -i '' "s/$required_version_regex/$required_version_replacement/g" "$versions_file"
  fi

  if grep -q "required_version" "$main_file" > /dev/null 2>&1; then
    echo "[WARN] Found legacy usage of required_version in '$main_file'."
    legacy_required_version_uses+=("$main_file")
  fi

  if grep -q 'when[[:space:]]*=[[:space:]]*"\?destroy"\?' "$main_file" > /dev/null 2>&1; then
    echo "[WARN] Found usage of destroy provisioner in '$main_file'."
    destroy_provisioner_uses+=("$main_file")
  fi
done

echo
echo "Next steps:"
echo

if [[ -n "${legacy_required_version_uses[*]}" ]]; then
  echo '=== required_version usage ==='
  echo "We now handle required_version in versions.tf, so you now need to remove any 'required_version' usage and related comments from the following files:"
  for file in "${legacy_required_version_uses[@]}"; do
    echo "- $file"
  done
  echo
fi

if [[ -n "${destroy_provisioner_uses[*]}" ]]; then
  echo "=== destroy provisioner usage ==="
  echo "Terraform 0.13 does not allow destroy-time provisioners to refer to other resources. Check the following files and fix if necessary. https://www.terraform.io/upgrade-guides/0-13.html#destroy-time-provisioners-may-not-refer-to-other-resources"
  for file in "${destroy_provisioner_uses[@]}"; do
    echo "- $file"
  done
  echo
fi

echo "=== Commit ==="
echo "Once all the above is done, do the following:"
echo "- Check the diffs in Git"
echo "- Test the code"
echo "- Submit a PR"