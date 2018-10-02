package main

import (
	"os"
	"reflect"
	"sort"
	"testing"
)

var tabbedOutputTests = []struct {
	in  string
	out bool
}{
	// Leading spaces are tabbed
	{"    1.21 Gigawatts!", true},
	{"\tGreat Scott!", true},
	{"\nRoads? Where we're going, we don't need roads.", true},
	// ... trailing spaces are not tabbed
	{"Whoa. This is heavy.\t", false},
	{"Ah... Are you telling me that you built a time machine... out of a DeLorean?\n", false},
	{"If you put your mind to it, you can accomplish anything.    ", false},
	// ... and no spaces are not tabbed
	{"I finally invent something that works!", false},
}

func TestOutputIsTabbed(t *testing.T) {
	for _, tt := range tabbedOutputTests {
		t.Run(tt.in, func(t *testing.T) {
			result := outputIsTabbed(tt.in)
			if result != tt.out {
				t.Errorf("Expected check to return %t, got %t", tt.out, result)
			}
		})
	}
}

func TestParseTestOutputFindsFailedTestCases(t *testing.T) {
	file, err := os.Open("./test_data/includes_failed_test.log")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	_, failedTests := parseTestOutput(file)
	expectedFailedTests := map[string]bool{
		"TestAlbDockerServiceWithAutoScaling": true,
	}
	if !reflect.DeepEqual(failedTests, expectedFailedTests) {
		t.Fatal("Expected to only find one test failed")
	}
}

func TestParseTestOutputFindsAllTestCases(t *testing.T) {
	file, err := os.Open("./test_data/includes_failed_test.log")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	testCaseOutput, _ := parseTestOutput(file)
	expectedTests := []string{
		"TestDockerFargateServiceWithAlb",
		"TestDeployEcsTask",
		"TestDockerFargateServiceWithNlb",
		"TestAlbDockerServiceWithAutoScaling",
	}
	foundTests := []string{}
	for testName, _ := range testCaseOutput {
		foundTests = append(foundTests, testName)
	}
	sort.Strings(expectedTests)
	sort.Strings(foundTests)
	if !reflect.DeepEqual(foundTests, expectedTests) {
		t.Fatal("Did not find all tests expected in log")
	}
}

func TestParseTestOutputCollectsOutputForPassedTest(t *testing.T) {
	file, err := os.Open("./test_data/basic_example.log")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	testCaseOutput, _ := parseTestOutput(file)
	expectedOutput := "TestTimeMachine 1985/10/26 01:18:00 Einstein enters time machine\nTestTimeMachine 1985/10/26 01:20:00 Einstein arrives from the past\n"
	if testCaseOutput["TestTimeMachine"] != expectedOutput {
		t.Fatal("Did not combine the interleaved output correctly")
	}
}

func TestParseTestOutputCollectsOutputForFailedTest(t *testing.T) {
	file, err := os.Open("./test_data/basic_example.log")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	testCaseOutput, _ := parseTestOutput(file)
	expectedOutput := "TestPlutonium 1985/10/12 00:00:00 Found plutonium\n--- FAIL: TestPlutonium (14d)\n    could not get bomb from Doctor Emmett Brown\n"
	if testCaseOutput["TestPlutonium"] != expectedOutput {
		t.Fatal("Did not combine the interleaved output correctly")
	}
}
