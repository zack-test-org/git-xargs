terraform {
  required_providers {
    secrets = {
      versions = ["0.1"]
      source   = "gruntwork.io/prototypes/secrets"
    }
  }
}

provider "secrets" {
  username = "username-test"
  password = "password-test"
}

//data "secrets_value" "example" {
//  name = "foo"
//}

resource "secrets_value" "example" {
  name  = "foo"
  value = "secret-value"
}

//output "read_secret" {
//  value = data.secrets_value.example
//}

output "written_secret" {
  value = secrets_value.example
}