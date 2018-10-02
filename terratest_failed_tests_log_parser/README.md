# Terratest Failed Tests Log Parser

A parser that will format the output of a parallel terratest log to only show
the results for the failed tests.

## Usage

- `go build terratest_failed_tests_log_parser.go`
- Download output from CircleCI (for example, https://circleci.com/gh/gruntwork-io/module-ecs/615)
- `./terratest_failed_tests_log_parser output.txt`
