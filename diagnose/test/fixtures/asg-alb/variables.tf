variable "server_port" {
  description = "The port the server will use for HTTP requests"
  type        = number
  default     = 8080
}

variable "alb_name" {
  description = "The name of the ALB"
  type        = string
  default     = "terraform-asg-example"
}

variable "instance_security_group_name" {
  description = "The name of the security group for the EC2 Instances"
  type        = string
  default     = "terraform-example-instance"
}

variable "alb_security_group_name" {
  description = "The name of the security group for the ALB"
  type        = string
  default     = "terraform-example-alb"
}

variable "iam_role_name" {
  description = "The name of the IAM role and instance profile"
  type        = string
  default     = "terraform-asg-example"
}

variable "ssh_key_name" {
  description = "The SSH Key Pair to associate with each EC2 instance"
  type        = string
  default     = "jim-brikman"
}