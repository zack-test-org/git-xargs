terraform {
  required_version = ">= 0.12"
}

# ---------------------------------------------------------------------------------------------------------------------
# CONFIGURE OUR AWS AND KUBERNETES CONNECTION
# ---------------------------------------------------------------------------------------------------------------------

provider "aws" {
  region = var.aws_region
}

provider "kubernetes" {
  version = "~> 1.6"

  load_config_file       = false
  host                   = data.aws_eks_cluster.cluster.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.kubernetes_token.token
}

data "aws_eks_cluster" "cluster" {
  name = var.eks_cluster_name
}

data "aws_eks_cluster_auth" "kubernetes_token" {
  name = var.eks_cluster_name
}


# ---------------------------------------------------------------------------------------------------------------------
# CREATE DYNAMODB TABLE FOR STORING BENCHMARK RESULTS
# ---------------------------------------------------------------------------------------------------------------------

resource "aws_dynamodb_table" "cpu_benchmark_results" {
  name         = "EKSFargateCPUBenchmark"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "UUID"

  attribute {
    name = "UUID"
    type = "S"
  }
}


# ---------------------------------------------------------------------------------------------------------------------
# CREATE SERVICE ACCOUNT WITH IAM ROLE WITH ACCESS TO DYNAMODB
# ---------------------------------------------------------------------------------------------------------------------

resource "kubernetes_service_account" "cpu_benchmark" {
  depends_on = [aws_iam_role_policy_attachment.cpu_benchmark]

  metadata {
    name      = "cpu-benchmark"
    namespace = "default"
    annotations = {
      "eks.amazonaws.com/role-arn" = aws_iam_role.cpu_benchmark.arn
    }
  }

  automount_service_account_token = true
}

resource "aws_iam_role" "cpu_benchmark" {
  name               = "${var.eks_cluster_name}-cpu-benchmark"
  assume_role_policy = module.service_account_assume_role_policy.assume_role_policy_json
}

resource "aws_iam_policy" "cpu_benchmark" {
  name   = "${var.eks_cluster_name}-cpu-benchmark-dynamodb-iam-policy"
  policy = data.aws_iam_policy_document.cpu_benchmark.json
}

resource "aws_iam_role_policy_attachment" "cpu_benchmark" {
  role       = aws_iam_role.cpu_benchmark.name
  policy_arn = aws_iam_policy.cpu_benchmark.arn
}

data "aws_iam_policy_document" "cpu_benchmark" {
  statement {
    actions   = ["dynamodb:PutItem"]
    resources = [aws_dynamodb_table.cpu_benchmark_results.arn]
  }
}

module "service_account_assume_role_policy" {
  source = "git::git@github.com:gruntwork-io/terraform-aws-eks.git//modules/eks-iam-role-assume-role-policy-for-service-account?ref=v0.11.1"

  eks_openid_connect_provider_arn = var.eks_openid_connect_provider_arn
  eks_openid_connect_provider_url = var.eks_openid_connect_provider_url
  namespaces                      = []
  service_accounts = [{
    name      = "cpu-benchmark"
    namespace = "default"
  }]
}


# ---------------------------------------------------------------------------------------------------------------------
# RUN BENCHMARK
# ---------------------------------------------------------------------------------------------------------------------

resource "kubernetes_config_map" "cpu_benchmark" {
  metadata {
    name      = "cpu-benchmark-scripts"
    namespace = "default"
  }

  data = {
    "entrypoint.sh" = file("${path.module}/entrypoint.sh")
    "benchmark.py"  = file("${path.module}/benchmark.py")
  }

  binary_data = {
    "data.json.gz" = filebase64("${path.module}/data.json.gz")
  }
}

resource "kubernetes_job" "cpu_benchmark" {
  depends_on = [
    kubernetes_config_map.cpu_benchmark,
  ]

  metadata {
    name      = "cpu-benchmark"
    namespace = "default"
  }

  spec {
    template {
      metadata {}
      spec {
        service_account_name = kubernetes_service_account.cpu_benchmark.metadata[0].name

        container {
          name    = "python"
          image   = "python:3.8.1-alpine3.11"
          command = ["/bin/sh", "/tmp/scripts/entrypoint.sh"]

          volume_mount {
            name       = "scripts-volume"
            read_only  = true
            mount_path = "/tmp/scripts"
          }
          # Make sure to mount the ServiceAccount token.
          volume_mount {
            mount_path = "/var/run/secrets/kubernetes.io/serviceaccount"
            name       = kubernetes_service_account.cpu_benchmark.default_secret_name
            read_only  = true
          }

          env {
            name  = "BENCHMARK_TABLE_NAME"
            value = aws_dynamodb_table.cpu_benchmark_results.name
          }
          env {
            name  = "REGION"
            value = var.aws_region
          }
        }
        volume {
          name = "scripts-volume"
          config_map {
            name         = "cpu-benchmark-scripts"
            default_mode = "0755"
          }
        }
        # We have to mount the service account token so that Tiller can access the Kubernetes API as the attached
        # ServiceAccount.
        volume {
          name = kubernetes_service_account.cpu_benchmark.default_secret_name

          secret {
            secret_name = kubernetes_service_account.cpu_benchmark.default_secret_name
          }
        }
        restart_policy = "Never"
      }
    }
    completions = var.num_instances
    parallelism = var.parallelism
  }
}
