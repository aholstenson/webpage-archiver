package captures

import (
	"github.com/aholstenson/webpage-archiver/pkg/outputs"
	"github.com/aholstenson/webpage-archiver/pkg/progress"
)

type capturerOptions struct {
	reporter  progress.Reporter
	output    outputs.Output
	userAgent string

	screenshotDirectory string
	screenshotPrefix    string
}

type Option func(o *capturerOptions)

func WithReporter(reporter progress.Reporter) Option {
	return func(o *capturerOptions) {
		o.reporter = reporter
	}
}

func WithOutput(output outputs.Output) Option {
	return func(o *capturerOptions) {
		o.output = output
	}
}

func WithUserAgent(ua string) Option {
	return func(o *capturerOptions) {
		o.userAgent = ua
	}
}

func WithScreenshots(directory string, prefix string) Option {
	return func(o *capturerOptions) {
		o.screenshotDirectory = directory
		o.screenshotPrefix = prefix
	}
}
