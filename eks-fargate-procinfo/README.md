# EKS Fargate Processor Info

This folder contains a very crude benchmark test you can run against EKS Fargate to get some insight into the underlying
hardware used for Fargate.

## Usage

1. Fill out `terraform.tfvars`
1. Run `go test -timeout 2h .`
