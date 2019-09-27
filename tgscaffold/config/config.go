package config

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/BurntSushi/toml"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/gruntwork-cli/files"
	"github.com/mitchellh/go-homedir"
)

type TgScaffoldConfig struct {
	InfrastructureModulesRepo           string
	InfrastructureModulesLocalPath      string
	DefaultInfrastructureModulesVersion string
}

func ParseConfig(configPath string) (TgScaffoldConfig, error) {
	var config TgScaffoldConfig

	configPathExpanded, err := homedir.Expand(configPath)
	if err != nil {
		return config, errors.WithStackTrace(err)
	}

	configData, err := files.ReadFileAsString(configPathExpanded)
	if err != nil {
		return config, errors.WithStackTrace(err)
	}

	_, err = toml.Decode(configData, &config)
	return config, errors.WithStackTrace(err)
}

func InitializeConfig(configPath string) error {
	prompts := []*survey.Question{
		{
			Name:     "InfrastructureModulesRepo",
			Prompt:   &survey.Input{Message: "Enter the git URL of your infrastructure-modules repository:"},
			Validate: survey.Required,
		},
		{
			Name:     "InfrastructureModulesLocalPath",
			Prompt:   &survey.Input{Message: "Enter the path to where the infrastructure-modules repository is cloned:"},
			Validate: survey.Required,
		},

		{
			Name:     "DefaultInfrastructureModulesVersion",
			Prompt:   &survey.Input{Message: "Enter the default version of your infrastructure-modules repository you would like to use:"},
			Validate: survey.Required,
		},
	}

	var initialConfig TgScaffoldConfig
	err := survey.Ask(prompts, &initialConfig)
	if err != nil {
		return errors.WithStackTrace(err)
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return errors.WithStackTrace(err)
	}
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(initialConfig); err != nil {
		return errors.WithStackTrace(err)
	}
	return errors.WithStackTrace(ioutil.WriteFile(configPath, buf.Bytes(), 0644))
}
