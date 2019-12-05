package main

import (
	"github.com/gruntwork-io/gruntwork-cli/entrypoint"
	"github.com/urfave/cli"

	"github.com/gruntwork-io/prototypes/tgscaffold/cmd"
)

// This variable is set at build time using -ldflags parameters. For example, we typically set this flag in circle.yml
// to the latest Git tag when building our Go apps:
//
// build-go-binaries --app-name my-app --dest-path bin --ld-flags "-X main.VERSION=$CIRCLE_TAG"
//
// For more info, see: http://stackoverflow.com/a/11355611/483528
var VERSION string

// main should only setup the CLI flags and help texts.
func main() {
	app := entrypoint.NewApp()
	entrypoint.HelpTextLineWidth = 120

	// Override the CLI FlagEnvHinter so it only returns the Usage text of the Flag and doesn't apend the envVar text. Original func https://github.com/urfave/cli/blob/master/flag.go#L652
	cli.FlagEnvHinter = func(envVar, str string) string {
		return str
	}

	app.Name = "tgscaffold"
	app.Author = "Gruntwork <www.gruntwork.io>"
	app.Description = "A CLI application for scaffolding terragrunt configuration"
	app.EnableBashCompletion = true
	// Set the version number from your app from the VERSION variable that is passed in at build time
	app.Version = VERSION

	app.Flags = cmd.GlobalFlags
	app.Commands = []cli.Command{
		cmd.GetComponentCommand(),
	}
	entrypoint.RunApp(app)
}
