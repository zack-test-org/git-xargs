# ---------------------------------------------------------------------------------------------------------------------
# ENVIRONMENT VARIABLES
# Define these secrets as environment variables
# ---------------------------------------------------------------------------------------------------------------------
{{- if eq .Cloud .Constants.AWS }}

# AWS_ACCESS_KEY_ID
# AWS_SECRET_ACCESS_KEY
{{- else if eq .Cloud .Constants.GCP }}

# GOOGLE_CLOUD_PROJECT
# GOOGLE_APPLICATION_CREDENTIALS
{{- end }}

# ---------------------------------------------------------------------------------------------------------------------
# TERRAGRUNT CONFIGURATION
# This is the configuration for Terragrunt, a thin wrapper for Terraform that supports locking and enforces best
# practices: https://github.com/gruntwork-io/terragrunt
# ---------------------------------------------------------------------------------------------------------------------

# Terragrunt will copy the Terraform configurations specified by the source parameter, along with any files in the
# working directory, into a temporary folder, and execute your Terraform commands in that folder.
terraform {
  source = "{{ .InfrastructureModulesSource }}//{{ .PathToModule }}?ref={{ .InfrastructureModulesVersion }}"
}

# Include all settings from the root terragrunt.hcl file
include {
  path = find_in_parent_folders()
}

# ---------------------------------------------------------------------------------------------------------------------
# MODULE PARAMETERS
# These variables are expected to be passed in by the operator
# ---------------------------------------------------------------------------------------------------------------------

inputs = {
  # -------------------------------------------------------------------------------------------------------------------
  # REQUIRED VARIABLES
  # The following input variables are required
  # -------------------------------------------------------------------------------------------------------------------

  {{- range .RequiredVars }}

  # {{ .Description }}
  # Type: {{ .Type }}
  {{ .Name }} = nil
  {{- end }}

  # -------------------------------------------------------------------------------------------------------------------
  # OPTIONAL VARIABLES
  # The following input variables are optional and have reasonable defaults
  # -------------------------------------------------------------------------------------------------------------------

  {{- range .OptionalVars }}

  # {{ .Description }}
  # Type: {{ .Type }}
  # {{ .Name }} = {{ .Default }}
  {{- end }}
}
