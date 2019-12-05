package main

import (
	"encoding/json"
	"fmt"
	"github.com/gruntwork-io/prototypes/drifter/util"
	"log"
	"os"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:     "module-dir, D",
			Value:    "current",
			Usage:    "Directory where terraform modules are expected to be found",
			Required: false,
		},
		cli.StringFlag{
			Name:     "github-repository, G",
			Usage:    "Git checkout URL",
			Required: false,
		},
		cli.BoolFlag{
			Name:     "recursive, R",
			Usage:    "Check modules recursively",
			Required: false,
		},
	}

	app.Action = func(c *cli.Context) error {
		directory := c.String("module-dir")
		recursive := c.Bool("recursive")

		repo := c.String("github-repository")

		if len(repo) > 0 {
			tmpDir, err := util.CheckoutRepoToTempFolder(repo)
			if err != nil {
				return err
			}
			directory = tmpDir
			recursive = true
		} else {
			if directory == "current" {
				currentDir, err := os.Getwd()
				if err != nil {
					return err
				}

				directory = currentDir
			}
		}

		log.Printf("Dir is %s", directory)

		log.Printf("Recursive is %v", recursive)

		dirs := util.ListDirectories(directory, recursive)

		endResult := map[string][]string{}

		for _, dir := range dirs {
			module, diags := tfconfig.LoadModule(dir)

			if diags != nil && diags.HasErrors() {
				return diags.Err()
			}

			for _, v := range module.ModuleCalls {
				moduleConfig, err := util.ParseRepository(v.Source)
				if err != nil {
					return err
				}
				if moduleConfig.NeedsUpdate {
					if _, ok := endResult[dir]; !ok {
						endResult[dir] = []string{}
					}
					endResult[dir] = append(endResult[dir], fmt.Sprintf("%v has available update from %v to %v", moduleConfig.Repository, moduleConfig.CurrentRelease, moduleConfig.LatestRelease))
				}
			}

		}

		b, err := json.MarshalIndent(endResult, "", "  ")
		if err == nil {
			fmt.Println(string(b))
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
