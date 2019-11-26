# ---------------------------------------------------------------------------------------------------------------------
# DEPLOY AN ASG + ELB + ROUTE 53 ENTRY THAT CAN BE USED FOR TESTING
# ---------------------------------------------------------------------------------------------------------------------

terraform {
  required_version = ">= 0.12, < 0.13"
}

provider "aws" {
  region = "us-east-2"

  # Allow any 2.x version of the AWS provider
  version = "~> 2.0"
}

# ---------------------------------------------------------------------------------------------------------------------
# CONFIGURE THE LAUNCH CONFIGURATION AND AUTO SCALING GROUP
# ---------------------------------------------------------------------------------------------------------------------

locals {
  working_user_data = <<-EOF
                      #!/bin/bash
                      echo "Hello, World" > index.html
                      nohup busybox httpd -f -p ${var.server_port} &
                      EOF

  broken_user_data = <<-EOF
                     #!/bin/bash
                     echo "This is an example of a broken User Data script that will fail to start a web server."
                     exit 1
                     EOF
}

resource "aws_launch_configuration" "example" {
  name_prefix          = var.name
  image_id             = data.aws_ami.ubuntu.image_id
  instance_type        = "t2.micro"
  security_groups      = [aws_security_group.instance.id]
  iam_instance_profile = aws_iam_instance_profile.instance.name
  key_name             = var.ssh_key_name
  user_data            = var.enable_broken_user_data ? local.broken_user_data : local.working_user_data

  # Required when using a launch configuration with an auto scaling group.
  # https://www.terraform.io/docs/providers/aws/r/launch_configuration.html
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "example" {
  launch_configuration = aws_launch_configuration.example.name
  vpc_zone_identifier  = data.aws_subnet_ids.default.ids

  target_group_arns = [aws_lb_target_group.asg.arn]
  health_check_type = "ELB"

  min_size = var.num_instances
  max_size = var.num_instances

  tag {
    key                 = "Name"
    value               = var.name
    propagate_at_launch = true
  }
}

# ---------------------------------------------------------------------------------------------------------------------
# CONFIGURE AN IAM ROLE
# We include SSM permissions in the IAM role so we can execute commands on the EC2 instances via SSM
# ---------------------------------------------------------------------------------------------------------------------

resource "aws_iam_instance_profile" "instance" {
  name = var.name
  role = aws_iam_role.instance.name
}

resource "aws_iam_role" "instance" {
  name               = var.name
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}


data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "ssm" {
  role   = aws_iam_role.instance.id
  policy = data.aws_iam_policy_document.ssm.json
}

data "aws_iam_policy_document" "ssm" {
  statement {
    effect = "Allow"
    actions = [
      "ssm:DescribeAssociation",
      "ssm:GetDeployablePatchSnapshotForInstance",
      "ssm:GetDocument",
      "ssm:DescribeDocument",
      "ssm:GetManifest",
      "ssm:GetParameter",
      "ssm:GetParameters",
      "ssm:ListAssociations",
      "ssm:ListInstanceAssociations",
      "ssm:PutInventory",
      "ssm:PutComplianceItems",
      "ssm:PutConfigurePackageResult",
      "ssm:UpdateAssociationStatus",
      "ssm:UpdateInstanceAssociationStatus",
      "ssm:UpdateInstanceInformation"
    ]
    resources = ["*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "ssmmessages:CreateControlChannel",
      "ssmmessages:CreateDataChannel",
      "ssmmessages:OpenControlChannel",
      "ssmmessages:OpenDataChannel"
    ]
    resources = ["*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "ec2messages:AcknowledgeMessage",
      "ec2messages:DeleteMessage",
      "ec2messages:FailMessage",
      "ec2messages:GetEndpoint",
      "ec2messages:GetMessages",
      "ec2messages:SendReply"
    ]
    resources = ["*"]
  }
}

# ---------------------------------------------------------------------------------------------------------------------
# CONFIGURE AN ALB TO ROUTE TRAFFIC ACROSS THE INSTANCES
# ---------------------------------------------------------------------------------------------------------------------

resource "aws_lb" "example" {
  name = var.name

  load_balancer_type = "application"
  subnets            = data.aws_subnet_ids.default.ids
  security_groups    = [aws_security_group.alb.id]
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.example.arn
  port              = var.alb_port
  protocol          = "HTTP"

  # By default, return a simple 404 page
  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "404: page not found"
      status_code  = 404
    }
  }
}

resource "aws_lb_target_group" "asg" {
  name = var.name

  port     = var.server_port
  protocol = "HTTP"
  vpc_id   = data.aws_vpc.default.id

  health_check {
    path                = "/"
    protocol            = "HTTP"
    matcher             = "200"
    interval            = 15
    timeout             = 3
    healthy_threshold   = 2
    unhealthy_threshold = 2
  }
}

resource "aws_lb_listener_rule" "asg" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 100

  condition {
    field  = "path-pattern"
    values = ["*"]
  }

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.asg.arn
  }
}

# ---------------------------------------------------------------------------------------------------------------------
# CONFIGURE THE SECURITY GROUPS FOR THE INSTANCES AND ALB
# ---------------------------------------------------------------------------------------------------------------------

resource "aws_security_group" "instance" {
  name = "${var.name}-instance"
}

# Allow inbound HTTP requests from the ALB... unless enable_broken_instance_security_group_settings is true, in
# which case this rule will not be included, and the ALB won't be able to talk to these instances.
resource "aws_security_group_rule" "allow_inbound_from_elb" {
  count                    = var.enable_broken_instance_security_group_settings ? 0 : 1
  type                     = "ingress"
  from_port                = var.server_port
  to_port                  = var.server_port
  protocol                 = "tcp"
  security_group_id        = aws_security_group.instance.id
  source_security_group_id = aws_security_group.alb.id
}

resource "aws_security_group_rule" "allow_all_outbound_instance" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  security_group_id = aws_security_group.instance.id
}

resource "aws_security_group" "alb" {
  name = "${var.name}-alb"
}

resource "aws_security_group_rule" "allow_http_inbound" {
  type              = "ingress"
  from_port         = var.alb_port
  to_port           = var.alb_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.alb.id
}

# Allow all outbound requests... unless enable_broken_elb_security_group_settings is true, in which case this rule
# will not be included, and the ALB won't be able to health check the instances.
resource "aws_security_group_rule" "allow_all_outbound" {
  count             = var.enable_broken_elb_security_group_settings ? 0 : 1
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.alb.id
}

# ---------------------------------------------------------------------------------------------------------------------
# CONFIGURE A ROUTE 53 RECORD FOR THE ALB
# ---------------------------------------------------------------------------------------------------------------------

resource "aws_route53_record" "alias" {
  name    = var.subdomain_name == null ? var.domain_name : "${var.subdomain_name}.${var.domain_name}"
  type    = "A"
  zone_id = data.aws_route53_zone.hosted_zone.zone_id

  alias {
    name                   = aws_lb.example.dns_name
    zone_id                = aws_lb.example.zone_id
    evaluate_target_health = false
  }
}

