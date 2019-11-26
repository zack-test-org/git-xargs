package main

import (
	"github.com/gruntwork-io/gruntwork-cli/entrypoint"
	"github.com/gruntwork-io/prototypes/diagnose/diagnose"
	"github.com/gruntwork-io/prototypes/diagnose/options"
	"github.com/urfave/cli"
)

// This variable is set at build time using -ldflags parameters. For example, we typically set this flag in circle.yml
// to the latest Git tag when building our Go apps:
//
// build-go-binaries --app-name my-app --dest-path bin --ld-flags "-X main.VERSION=$CIRCLE_TAG"
//
// For more info, see: http://stackoverflow.com/a/11355611/483528
var VERSION string

func main() {
	// Create a new CLI app. This will return a urfave/cli App with some
	// common initialization.
	app := entrypoint.NewApp()

	app.Name = "diagnose"
	app.Author = "Gruntwork <www.gruntwork.io>"

	// Set the version number from your app from the VERSION variable that is passed in at build time
	app.Version = VERSION

	app.Action = func(cliContext *cli.Context) error {
		opts, err := options.ParseOpts(cliContext)
		if err != nil {
			return err
		}

		return diagnose.Diagnose(opts)
	}

	// Run your app using the entrypoint package, which will take care of exit codes, stack traces, and panics
	entrypoint.RunApp(app)
}
