package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func outputIsTabbed(data string) bool {
	re := regexp.MustCompile("^\\s+.+")
	matched := re.MatchString(data)
	return matched
}

// parseTestOutput takes a file handle pointing to test output from a parallel
// go test and returns two maps:
// - testCaseOutput that maps test names to all the log entries for that test
//   case as a string
// - failedTests is a set (implemented as a map) that indicates which tests have failed.
func parseTestOutput(file *os.File) (testCaseOutput map[string]string, failedTests map[string]bool) {
	testCaseOutput = make(map[string]string)
	failedTests = make(map[string]bool)
	parsingFailedTest := ""
	re := regexp.MustCompile(`--- FAIL: (?P<test>Test[^\s]+) `)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Text()

		// Mini state machine: The parser has 2 states:
		// - general parsing
		// - parsing a failed test
		// In the second state, we need to agglomerate all the stack traces
		// under the relevant test, so we keep track when we enter a fail block
		// using parsingFailedTest, exiting when we enter the blank newline
		// that terminates failed test output.
		if parsingFailedTest != "" && outputIsTabbed(data) {
			// agglomerate output to the respective test case, as this is part
			// of the error stack trace for a test failure
			testCaseOutput[parsingFailedTest] += data + "\n"
		} else if parsingFailedTest != "" && strings.TrimSpace(data) == "" {
			// Exit fail test parsing mode
			parsingFailedTest = ""
		} else if strings.HasPrefix(data, "Test") {
			vals := strings.Split(data, " ")
			testCaseOutput[vals[0]] += data + "\n"
		} else if strings.HasPrefix(data, "--- FAIL") {
			m := re.FindStringSubmatch(data)
			testCaseOutput[m[1]] += data + "\n"
			failedTests[m[1]] = true
			parsingFailedTest = m[1]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return testCaseOutput, failedTests
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	testCaseOutput, failedTests := parseTestOutput(file)

	for key, _ := range failedTests {
		fmt.Println(key)
		fmt.Println("------------")
		fmt.Println(testCaseOutput[key])
		fmt.Println("")
	}
}
