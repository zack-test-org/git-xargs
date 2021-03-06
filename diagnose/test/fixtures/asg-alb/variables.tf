variable "enable_broken_user_data" {
  description = "Set this flag to true to launch the EC2 Instances with broken user data that doesn't successfully start a web server in the EC2 Instances. This is useful to test if the diagnose utility can figure out that the problem is with the code on the server itself."
  type        = bool
  default     = false
}

variable "enable_broken_instance_security_group_settings" {
  description = "Set this flag to true to misconfigure the Security Group for the EC2 Instances so the ELB can't talk to them. This is useful to test if the diagnose utility can figure out that the problem is with the Security Group of the instances."
  type        = bool
  default     = false
}

variable "enable_broken_elb_security_group_settings" {
  description = "Set this flag to true to misconfigure the Security Group for the ELB so it can't make outbound requests, including health checks. This is useful to test if the diagnose utility can figure out that the problem is with the Security Group of the ELB."
  type        = bool
  default     = false
}

variable "name" {
  description = "The name to use for all the resources created by this module"
  type        = string
  default     = "terraform-asg-example"
}

variable "server_port" {
  description = "The port the server will use for HTTP requests"
  type        = number
  default     = 8080
}

variable "alb_port" {
  description = "The port the ALB will use for HTTP requests"
  type        = number
  default     = 80
}

variable "ssh_key_name" {
  description = "The SSH Key Pair to associate with each EC2 instance"
  type        = string
  default     = null
}

variable "domain_name" {
  description = "The domain name used to find the Route 53 Hosted Zone onto which a Route 53 alias record will be added"
  type        = string
  default     = "gruntwork.in"
}

variable "domain_tags" {
  description = "Optional tags usued to further filter to the right Route 53 Hosted Zone specified by var.dommain_name"
  type        = map(string)
  default     = {
    original = "true"
  }
}

variable "subdomain_name" {
  description = "The subdomain of var.domain_name in which to create the Route 53 alias record. If null, the alias will point directly to var.domain_name."
  type        = string
  default     = "jimtest"
}

variable "num_instances" {
  description = "The number of Instances to run in the ASG"
  type        = number
  default     = 2
}

