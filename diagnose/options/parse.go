package options

import (
	"fmt"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/urfave/cli"
)

func ParseOpts(cliContext *cli.Context) (*Options, error) {
	url := cliContext.Args().First()
	if url == "" {
		return nil, errors.WithStackTrace(MissingArgument("URL"))
	}

	logger := logging.GetLogger("diagnose")

	return &Options{
		Url:    url,
		Logger: logger,
	}, nil
}

type MissingArgument string

func (arg MissingArgument) Error() string {
	return fmt.Sprintf("Missing argument '%s'", string(arg))
}
