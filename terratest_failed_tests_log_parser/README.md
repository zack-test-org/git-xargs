# Terratest Failed Tests Log Parser

A parser that will format the output of a parallel terratest log to only show
the results for the failed tests.


## Usage

- `go build terratest_failed_tests_log_parser.go`
- Download output from CircleCI (for example, https://circleci.com/gh/gruntwork-io/module-ecs/615)
- `./terratest_failed_tests_log_parser output.txt`

This will then output the logs from each failed test
together, separating each one with a header containing the
test name.

See `test_data` directory for some examples. The files
ending with `_parsed.log` is a sample output given the
original file with the same prefix.
