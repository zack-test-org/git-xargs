terraform {
  required_providers {
    securesecrets = {
      versions = ["0.1"]
      source   = "gruntwork.io/prototypes/securesecrets"
    }
  }
}

provider "securesecrets" {
  region = "eu-west-1"
}

resource "securesecrets_value" "example" {
  name        = "jim-testing-terraform-provider"
  description = "This is Jim testing a new Terraform provider he created for securely storing secrets in AWS Secrets Manager"
  version     = "v2"
}

output "secret" {
  value = securesecrets_value.example
}