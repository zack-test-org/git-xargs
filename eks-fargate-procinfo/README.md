# EKS Fargate Processor Info

This folder contains a very crude benchmark test you can run against EKS Fargate to get some insight into the underlying
hardware used for Fargate.

## Usage

1. Fill out `terraform.tfvars`
1. Run `go test -timeout 2h .`



## TODO!!!!!

- Flush out README more, especially with motivation
- Add module for EKS cluster to make test easier to run
- Update test to output graphs and store as artifacts
- Add a test runner that will run across all regions
- Add comments to code everywhere
- Record video to add to hackfest video
