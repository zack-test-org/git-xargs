package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/urfave/cli"

	"github.com/gruntwork-io/prototypes/tgscaffold/config"
	"github.com/gruntwork-io/prototypes/tgscaffold/generator"
	"github.com/gruntwork-io/prototypes/tgscaffold/selector"
	"github.com/gruntwork-io/prototypes/tgscaffold/terraform"
)

var (
	outputDirFlag = cli.StringFlag{
		Name:  "output-dir",
		Usage: "Where to render the terragrunt.hcl file.",
		Value: ".",
	}
)

func GetComponentCommand() cli.Command {
	return cli.Command{
		Name:  "component",
		Usage: "Generate terragrunt config for deploying selected component.",
		Description: `This will scaffold a new terragrunt.hcl file that you can use in your infrastructure-live folder for deploying a component that lives in your configured infrastructure-modules repository. The terragrunt.hcl file will be seeded with all the input variables of the selected component, highlighting those that are required and thus need to be filled in.

This command will start by introspecting your infrastructure-modules directory and looking for deployable modules (folders that have a main.tf file). This command will then ask you to select one of those components, from which point it will parse the variables info and render a template of the terragrunt.hcl file.`,
		Flags: append(
			GlobalFlags,
			outputDirFlag,
		),
		Action: withInitialization(errors.WithPanicHandling(generateComponent)),
	}
}

func generateComponent(c *cli.Context) error {
	configPath := c.String(configPathFlag.Name)
	tgScaffoldConfig, err := config.ParseConfig(configPath)
	if err != nil {
		return err
	}

	selectedModule, err := selector.SelectModule(tgScaffoldConfig.InfrastructureModulesLocalPath)
	if err != nil {
		return err
	}
	selectedModuleFullPath := filepath.Join(tgScaffoldConfig.InfrastructureModulesLocalPath, selectedModule)

	variables, err := terraform.GetModuleVariables(selectedModuleFullPath)
	if err != nil {
		return err
	}
	requiredVars := []terraform.Variable{}
	optionalVars := []terraform.Variable{}
	for _, variable := range variables {
		variable.Type = strings.ReplaceAll(variable.Type, "\n", "\n# ")
		if variable.Default == nil {
			requiredVars = append(requiredVars, variable)
		} else {
			tmp := strings.ReplaceAll(*variable.Default, "\n", "\n# ")
			variable.Default = &tmp
			optionalVars = append(optionalVars, variable)
		}
	}
	renderOptions := generator.RenderOptions{
		Constants:                    generator.Constants,
		Cloud:                        generator.AWS,
		InfrastructureModulesSource:  fmt.Sprintf("git::ssh://%s", tgScaffoldConfig.InfrastructureModulesRepo),
		PathToModule:                 selectedModule,
		InfrastructureModulesVersion: tgScaffoldConfig.DefaultInfrastructureModulesVersion,
		RequiredVars:                 requiredVars,
		OptionalVars:                 optionalVars,
	}
	return generator.Render(renderOptions, c.String(outputDirFlag.Name))
}
