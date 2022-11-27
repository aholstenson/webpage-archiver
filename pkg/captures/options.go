package captures

import (
	"github.com/aholstenson/webpage-archiver/pkg/outputs"
	"github.com/aholstenson/webpage-archiver/pkg/progress"
)

type capturerOptions struct {
	reporter  progress.Reporter
	output    outputs.Output
	userAgent string
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
