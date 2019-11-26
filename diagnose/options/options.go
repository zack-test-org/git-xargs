package options

import (
	"github.com/sirupsen/logrus"
)

type Options struct {
	Url    string
	Logger *logrus.Logger
}
