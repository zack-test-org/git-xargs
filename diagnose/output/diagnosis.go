package output

import "github.com/gruntwork-io/prototypes/diagnose/options"

func ShowDiagnosis(diagnosis string, opts *options.Options) {
	opts.Logger.Infof("\n\n======= DIAGNOSIS =======\n\n%s\n\n", diagnosis)
}