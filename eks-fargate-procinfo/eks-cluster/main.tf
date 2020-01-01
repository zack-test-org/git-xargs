# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
# DEPLOY A EKS CLUSTER WITH FARGATE AS WORKERS
# These templates show an example of how to provision an EKS cluster set up to support Fargate to run workloads.
# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

terraform {
  # This module has been updated with 0.12 syntax, which means it is no longer compatible with any versions below 0.12.
  required_version = ">= 0.12"
}

# ------------------------------------------------------------------------------
# CONFIGURE OUR AWS CONNECTION
# ------------------------------------------------------------------------------

provider "aws" {
  # The fargate profile resource was introduced in AWS provider version 2.41
  version = "~> 2.41"

  region = var.aws_region
}

# ---------------------------------------------------------------------------------------------------------------------
# CREATE A VPC WITH NAT GATEWAY
# We will provision a new VPC because EKS requires a VPC that is tagged to denote that it is shared with Kubernetes.
# Specifically, both the VPC and the subnets that EKS resides in need to be tagged with:
# kubernetes.io/cluster/EKS_CLUSTER_NAME=shared
# This information is used by EKS to allocate ip addresses to the Kubernetes pods.
# ---------------------------------------------------------------------------------------------------------------------

module "vpc_app" {
  source = "git::git@github.com:gruntwork-io/module-vpc.git//modules/vpc-app?ref=yori-disable-endpoint"

  vpc_name   = var.vpc_name
  aws_region = var.aws_region

  # These tags (kubernetes.io/cluster/EKS_CLUSTERNAME=shared) are used by EKS to determine which AWS resources are
  # associated with the cluster. This information will ultimately be used by the [amazon-vpc-cni-k8s
  # plugin](https://github.com/aws/amazon-vpc-cni-k8s) to allocate ip addresses from the VPC to the Kubernetes pods.
  custom_tags = module.vpc_tags.vpc_eks_tags

  public_subnet_custom_tags              = module.vpc_tags.vpc_public_subnet_eks_tags
  private_app_subnet_custom_tags         = module.vpc_tags.vpc_private_app_subnet_eks_tags
  private_persistence_subnet_custom_tags = module.vpc_tags.vpc_private_persistence_subnet_eks_tags

  # The IP address range of the VPC in CIDR notation. A prefix of /18 is recommended. Do not use a prefix higher
  # than /27.
  cidr_block = "10.0.0.0/18"

  # The number of NAT Gateways to launch for this VPC. For production VPCs, a NAT Gateway should be placed in each
  # Availability Zone (so likely 3 total), whereas for non-prod VPCs, just one Availability Zone (and hence 1 NAT
  # Gateway) will suffice. Warning: You must have at least this number of Elastic IP's to spare.  The default AWS
  # limit is 5 per region, but you can request more.
  num_nat_gateways = 1
}

module "vpc_tags" {
  source = "git::git@github.com:gruntwork-io/terraform-aws-eks.git//modules/eks-vpc-tags?ref=yori-fargate"

  eks_cluster_names = [var.eks_cluster_name]
}

# ---------------------------------------------------------------------------------------------------------------------
# CREATE THE EKS CLUSTER IN TO THE VPC
# ---------------------------------------------------------------------------------------------------------------------

module "eks_cluster" {
  source = "git::git@github.com:gruntwork-io/terraform-aws-eks.git//modules/eks-cluster-control-plane?ref=yori-fargate"

  cluster_name       = var.eks_cluster_name
  kubernetes_version = var.kubernetes_version

  vpc_id                = module.vpc_app.vpc_id
  vpc_master_subnet_ids = local.usable_subnet_ids
  vpc_worker_subnet_ids = local.usable_subnet_ids

  configure_kubectl   = var.configure_kubectl
  kubectl_config_path = var.kubectl_config_path

  upgrade_cluster_script_wait_for_rollout = false

  # Make sure the cluster operates without worker nodes
  fargate_only = true
}

# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
# DATA SOURCES
# These resources must already exist.
# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

# Get the subnet ids while taking into account the availability zone whitelist.
data "aws_subnet" "all" {
  count = module.vpc_app.num_availability_zones
  id    = element(module.vpc_app.private_app_subnet_ids, count.index)
}

locals {
  # Get the list of availability zones to use for the cluster and node based on the whitelist.
  # Here, we use an awkward join and split because Terraform does not support conditional ternary expressions with list
  # values. See https://github.com/hashicorp/terraform/issues/12453
  availability_zones = split(
    ",",
    length(var.availability_zone_whitelist) == 0 ? join(",", data.aws_subnet.all.*.availability_zone) : join(",", var.availability_zone_whitelist),
  )

  usable_subnet_ids = matchkeys(
    data.aws_subnet.all.*.id,
    data.aws_subnet.all.*.availability_zone,
    local.availability_zones,
  )
}
