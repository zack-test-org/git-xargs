# ---------------------------------------------------------------------------------------------------------------------
# TO KEEP THIS EXAMPLE SIMPLE, DEPLOY INTO THE DEFAULT VPC & SUBNETS
# ---------------------------------------------------------------------------------------------------------------------

data "aws_vpc" "default" {
  default = true
}

data "aws_subnet_ids" "default" {
  vpc_id = data.aws_vpc.default.id
}

# ---------------------------------------------------------------------------------------------------------------------
# LOOK UP THE HOSTED ZONE
# ---------------------------------------------------------------------------------------------------------------------

data "aws_route53_zone" "hosted_zone" {
  name = var.domain_name
  tags = var.domain_tags
}

# ---------------------------------------------------------------------------------------------------------------------
# LOOK UP THE LATEST UBUNTU 18.04 AMI
# ---------------------------------------------------------------------------------------------------------------------

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "image-type"
    values = ["machine"]
  }

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-bionic-18.04-amd64-server-*"]
  }
}