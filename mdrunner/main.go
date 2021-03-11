package main

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const defaultLogLevel = "WARN"

// Snippet is a piece of executable shell code
type Snippet struct {
	Index            int      `yaml:"index"`
	Description      string   `yaml:"-"`
	Content          string   `yaml:"-"`
	State            string   `yaml:"state"`
	WorkingDirectory string   `yaml:"dir"`
	EnvVars          []string `yaml:"env"`
}

// getEnv returns either an environment variable or the specified default
func getEnv(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}

// setLogLevel sets the current log level depending on the MDRUNNER_LOGLEVEL environment variable
func setLogLevel(logLevel string) {
	logLevel = strings.ToUpper(logLevel)
	switch logLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	case "PANIC":
		log.SetLevel(log.PanicLevel)
	default:
		log.Fatal("Log level must be one of (DEBUG, INFO, WARN, ERROR, FATAL, PANIC)")
	}
}

// loadFile reads in the contents of a given filename to a string
func loadFile(filename string) (string, error) {

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// saveSnippet saves the contents of a string to a temporary file, making it executable
func saveSnippet(filename string, content string) error {
	log.Infof("Saving snippet to %s", filename)
	prefix := `#!/bin/bash
	
	`
	content = prefix + content + `

echo "### BEGIN mdrunner ###"
echo "# mdrunner:pwd $(pwd)"
IFS=$'\n'
for i in $(env) ; do
	echo "# mdrunner:env $i"
done
echo "### END mdrunner ###"
`
	return ioutil.WriteFile(filename, []byte(content), 0700)
}

// getDescription extracts the description of a step from the comments of the snippet
func getDescription(index int, content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimLeft(line, " \t")
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			line = strings.TrimPrefix(line, "#")
			line = strings.TrimLeft(line, " ")
			return fmt.Sprintf("Step %d: %s", index, line)
		} else {
			return fmt.Sprintf("Step %d", index)
		}
	}
	return fmt.Sprintf("Step %d", index)
}

// extractSnippets converts a string into an array of Snippets
func extractAllSnippets(snippets []Snippet, content string) ([]Snippet, error) {
	re := regexp.MustCompile("(?s)```bash\n(.+?)\n```\n")
	for i, single := range re.FindAllStringSubmatch(content, -1) {
		description := getDescription(i+1, single[1])
		if len(snippets) <= i {
			log.Debugf("Creating new snippet '%s' and appending to the array of snippets", description)
			snippet := Snippet{
				Index:       i + 1,
				Description: description,
				Content:     single[1],
				State:       "pending",
			}
			snippets = append(snippets, snippet)
		} else {
			log.Infof("Layering new description and content for '%s' from state loaded in", description)
			snippet := snippets[i]
			snippet.Description = description
			snippet.Content = single[1]
			snippets[i] = snippet
		}
	}
	return snippets, nil
}

// runSnippet executes a given Snippet and returns the output in an array of strings
func runSnippet(snippet Snippet, dir string, envVars []string) ([]string, error) {

	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	log.Infof("Changing directory to %s", dir)
	os.Chdir(dir)

	filename := fmt.Sprintf("/tmp/%d.sh", snippet.Index)
	err := saveSnippet(filename, snippet.Content)
	if err != nil {
		return nil, err
	}
	defer os.Remove(filename)

	cmd := exec.Command(filename)
	cmd.Env = append(os.Environ(), envVars...)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	var output []string
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		disableOutput := false
		for scanner.Scan() {
			line := scanner.Text()
			if line == "### BEGIN mdrunner ###" {
				disableOutput = true
			}
			output = append(output, line)
			if !disableOutput {
				fmt.Printf("%s\n", scanner.Text())
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}
	fmt.Println()
	return output, nil
}

// extractDir captures the working directory from the output of a command
func extractDir(text []string, dir string) string {
	// # mdrunner:pwd /Users/pete/git/mdrunner
	for _, line := range text {
		if strings.HasPrefix(line, "# mdrunner:pwd ") {
			dir := strings.TrimLeft(line, "# mdrunner:pwd ")
			log.Debugf("Capture working directory of %s", dir)
			return dir
		}
	}
	return dir
}

// extractEnvVars captures the environment variables from the output of a command
func extractEnvVars(text []string, originals []string) []string {
	// # mdrunner:env TERM_PROGRAM=iTerm.app
	envVars := []string{}
	for _, line := range text {
		if strings.HasPrefix(line, "# mdrunner:env ") {
			env := strings.TrimLeft(line, "# mdrunner:env ")
			skip := false
			for _, check := range originals {
				if env == check || strings.HasPrefix(env, "SHLVL=") || strings.HasPrefix(env, "_=") {
					skip = true
					break
				}
			}
			if !skip {
				log.Debugf("Capture environment variable: %s", env)
				envVars = append(envVars, env)
			}
		}
	}

	return envVars
}

func saveSnippetState(snippets []Snippet, filename string) error {
	log.Infof("Saving snippet state to %s", filename)
	y, err := yaml.Marshal(snippets)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, y, 0600)
}

// editSnippet pulls up an editor and edits the script prior to execution
func editSnippet(snippet Snippet) (Snippet, error) {

	tmpFile, err := ioutil.TempFile(os.TempDir(), "mdrunner-")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}

	if err := tmpFile.Close(); err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tmpFile.Name())

	err = ioutil.WriteFile(tmpFile.Name(), []byte(snippet.Content), 0600)
	if err != nil {
		return snippet, err
	}

	editor := getEnv("EDITOR", "vim")
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		return snippet, err
	}
	contents, err := loadFile(tmpFile.Name())
	if err != nil {
		return snippet, err
	}
	snippet.Content = contents
	return snippet, nil
}

// runAllSnippets sequentially runs through an array of Snippets
func runAllSnippets(snippets []Snippet, stateFileName string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	envVars := []string{}

	// Capture the environment variables from the beginning so that we can weed them out when storing state
	originalEnvVars := os.Environ()

	// var logLevel string
	// var resetState bool
	// var noPrompt bool
	// var verifyOnly bool
	// var runStep int
	// var onlyCleanup bool
	// var noCleanup bool

	// fmt.Printf("%#v\n", runSteps)
	// fmt.Printf("Total # of steps: %d\n", len(snippets))

	for i, snippet := range snippets {

		runThis := false
		if len(runSteps) > 0 {
			for _, k := range runSteps {
				if i+1 == k {
					runThis = true
					break
				}
			}
		}

		if len(runSteps) > 0 && !runThis {
			fmt.Printf("# Skipping %s\n", snippet.Description)
			continue
		}

		var finalAnswer string
		for answer := "start"; answer == "start" || answer == "edit"; {
			var qs = []*survey.Question{
				{
					Name: "proceed",
					Prompt: &survey.Select{
						Message: fmt.Sprintf("# %s", snippet.Description),
						Options: []string{"run", "skip", "edit", "quit"},
						Default: "run",
					},
				},
			}

			if i > 0 && snippets[i-1].WorkingDirectory != "" {
				log.Debugf("#### working dir: %s", snippets[i-1].WorkingDirectory)
				dir = snippets[i-1].WorkingDirectory
			}
			if i > 0 && len(snippets[i-1].EnvVars) > 0 {
				log.Debugf("##### env vars: %s", snippets[i-1].EnvVars)
				envVars = snippets[i-1].EnvVars
			}
			if snippet.State == "done" {
				answer = "skip"
			} else {

				fmt.Printf(`

############## %s
				
%s
				
##############################
				
				
`, snippet.Description, snippet.Content)

				if noPrompt {
					answer = "run"
				} else {
					// perform the questions
					err := survey.Ask(qs, &answer)
					if err != nil {
						fmt.Println(err.Error())
						return err
					}
				}

				if answer == "quit" {
					os.Exit(0)
				}

				if answer == "edit" {
					var err error
					snippet, err = editSnippet(snippet)
					if err != nil {
						return err
					}
				}
			}
			finalAnswer = answer
		}

		if finalAnswer == "skip" {
			fmt.Printf("# Skipping %s\n", snippet.Description)
			continue
		}

		fmt.Printf("# Running %s\n", snippet.Description)
		fmt.Println()

		output, err := runSnippet(snippet, dir, envVars)
		dir = extractDir(output, dir)
		envVars = extractEnvVars(output, originalEnvVars)

		snippet.WorkingDirectory = dir
		snippet.EnvVars = envVars
		if err != nil {
			snippet.State = "error"
			snippets[i] = snippet
			saveSnippetState(snippets, stateFileName)
			fmt.Printf("\nSaved state file %s\n\n", stateFileName)
			return err
		} else {
			snippet.State = "done"
			snippets[i] = snippet
			err := saveSnippetState(snippets, stateFileName)
			if err != nil {
				return err
			}
		}
	}
	os.Remove(stateFileName)
	return nil
}

func loadSnippetState(filename string) ([]Snippet, error) {
	log.Debugf("Detecting whether snippet state file %s exists", filename)
	if _, err := os.Stat(filename); err != nil {
		return nil, nil
	}
	fmt.Printf("Loading state from %s\n\n", filename)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var snippets []Snippet
	err = yaml.Unmarshal(yamlFile, &snippets)
	if err != nil {
		return nil, err
	}
	return snippets, nil
}

var rootCmd = &cobra.Command{
	Use:  "mdrunner FILE",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)

		var mdFile string
		var err error

		mdFile, err = filepath.Abs(args[0])
		if err != nil {
			log.Fatal(err)
		}
		if _, err := os.Stat(mdFile); err != nil {
			log.Fatal(errors.New("File does not exist"))
		}

		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(mdFile)))[0:7]
		hashFile := fmt.Sprintf("/tmp/mdrunner-%s.yaml", hash)

		stateFile := getEnv("MDRUNNER_STATEFILE", hashFile)

		if resetState {
			if _, err := os.Stat(stateFile); err == nil {
				log.Infof("Removing %s", stateFile)
				err := os.Remove(stateFile)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		snippets, err := loadSnippetState(stateFile)
		if err != nil {
			log.Fatal(err)
		}
		fileContent, err := loadFile(mdFile)
		if err != nil {
			log.Fatal(err)
		}
		snippets, err = extractAllSnippets(snippets, fileContent)
		if err != nil {
			log.Fatal(err)
		}
		if len(snippets) == 0 {
			log.Info("No snippets to run.")
		}

		if err := runAllSnippets(snippets, stateFile); err != nil {
			log.Fatal(err)
		}
	},
}

var logLevel string
var resetState bool
var noPrompt bool
var verifyOnly bool
var runSteps []int
var onlyCleanup bool
var noCleanup bool

func main() {
	logLevel = getEnv("MDRUNNER_LOGLEVEL", "")
	resetState = false
	noPrompt = false
	verifyOnly = false
	onlyCleanup = false
	onlyCleanup = false
	if logLevel == "" {
		rootCmd.Flags().StringVar(&logLevel, "log-level", defaultLogLevel, "Log level (DEBUG, INFO, WARN, ERROR, FATAL, PANIC) (can be set by MDRUNNER_LOGLEVEL)")
	} else {
		rootCmd.Flags().StringVar(&logLevel, "log-level", logLevel, "Currently set by MDRUNNER_LOGLEVEL")
	}
	rootCmd.Flags().BoolVar(&resetState, "reset", resetState, "Clear state prior to running (not implemented yet)")
	rootCmd.Flags().BoolVar(&noPrompt, "no-prompt", noPrompt, "Iterate over scripts without pausing (useful for automated testing)")
	rootCmd.Flags().BoolVar(&verifyOnly, "verify", verifyOnly, "Only run the steps marked by 'mdrunner:verify'")
	rootCmd.Flags().BoolVar(&noCleanup, "no-cleanup", noCleanup, "Prevent the cleanup step marked by 'mdrunner:cleanup' from running")
	rootCmd.Flags().BoolVar(&onlyCleanup, "cleanup", onlyCleanup, "Only run the steps marked by 'mdrunner:cleanup' (not implemented yet)")
	rootCmd.Flags().IntSliceVar(&runSteps, "step", nil, "Run only the specified step")

	rootCmd.Execute()
}

/*

show state file somehow

in markdown, add
# mdrunner:verify
# mdrunner:cleanup
in code:
--no-cleanup
--verify
--step 4
--step verify
--step cleanup
*/
