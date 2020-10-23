# Overview

Context converter is a CLI that allows you to quickly make mass-updates to your Github repositories' `.circleci/config.yml` files.

# How it works 

Currently, when you run the `context-converter`, you specify a Github organization name, such as `gruntwork-io`. The tool will: 

1. fetch all the public and private repositories owned by this organization 
1. filter down to only those repos containing a `.circleci/config.yml` file
1. ensure that the `.circleci/config.yml` `version` is `2.0` or greater, since context support 
1. add "Gruntwork Admin" to the `Workflows -> Jobs -> Contexts` arrays when necessary 

# Prerequisites 

The following binaries are **required** for the Context converter tool: 
* [yq](https://mikefarah.gitbook.io/yq/)

# Getting started 

1. Ensure you've installed the prerequisites! 
1. Create and export a Github personal access token 
```

export GITHUB_OAUTH_TOKEN
go run main.go
```

`GITHUB_OAUTH_TOKEN` must be a [Github personal access token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token) that was created with an account that is a member of the Gruntwork-io Github organization. 

# TODO

* Support all cases for context repair: 
- [x] Find `context` nodes and overwrite them (for testing purposes and familiarization with the yaml v3 API)
- [x] When no `context` node is present for a workflow job, add it along with the correct context
- [ ] When a `context` node is present, without the correct values, add the value
- [x] When a context node is present, with the correct value, do nothing
