package selector

import (
	"fmt"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-zglob"
)

func SelectModule(infraModulesPath string) (string, error) {
	moduleDirs, err := findModuleDirectories(infraModulesPath)
	if err != nil {
		return "", err
	}

	var selectedModule string
	prompt := &survey.Select{
		Message: "Which module would you like to add?: ",
		Options: moduleDirs,
	}
	err = survey.AskOne(prompt, &selectedModule, survey.WithValidator(survey.Required))
	return selectedModule, err
}

func findModuleDirectories(rootPath string) ([]string, error) {
	globPattern := fmt.Sprintf("%s/**/main.tf", rootPath)
	mainTfFiles, err := zglob.Glob(globPattern)
	if err != nil {
		return nil, err
	}
	moduleDirs := []string{}
	for _, mainTfPath := range mainTfFiles {
		moduleDir := filepath.Dir(mainTfPath)
		relativeModuleDir, err := filepath.Rel(rootPath, moduleDir)
		if err != nil {
			return nil, err
		}
		moduleDirs = append(moduleDirs, relativeModuleDir)
	}
	return moduleDirs, nil
}
