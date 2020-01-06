# ---------------------------------------------------------------------------------------------------------------------
# MODULE PARAMETERS
# These variables are expected to be passed in by the operator when calling this terraform module. These should match
# what is provided by the associated eks-cluster module.
# ---------------------------------------------------------------------------------------------------------------------

variable "aws_region" {
  description = "AWS Region where EKS cluster lives."
  type        = string
}

variable "eks_cluster_name" {
  description = "Name of the EKS cluster where to run the benchmark."
  type        = string
}

variable "eks_openid_connect_provider_arn" {
  description = "ARN of the OpenID Connect Provider for EKS to retrieve IAM credentials."
  type        = string
}

variable "eks_openid_connect_provider_url" {
  description = "URL of the OpenID Connect Provider for EKS to retrieve IAM credentials."
  type        = string
}

# ---------------------------------------------------------------------------------------------------------------------
# OPTIONAL MODULE PARAMETERS
# These variables have defaults, but may be overridden by the operator.
# ---------------------------------------------------------------------------------------------------------------------

variable "num_instances" {
  description = "Number of benchmark instances to provision."
  type        = number
  default     = 90
}

variable "parallelism" {
  description = "Number of pods to run in parallel at a time."
  type        = number
  default     = 90
}
