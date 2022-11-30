package archiver

import (
	"github.com/aholstenson/webpage-archiver/pkg/progress"
)

type archiverConfig struct {
	reporter  progress.Reporter
	userAgent string
}

type captureConfig struct {
	reporter       progress.Reporter
	userAgent      string
	screenshotFunc func([]byte) error
}

type Option interface {
	applyArchiver(o *archiverConfig)
}

type CaptureOption interface {
	applyCapture(o *captureConfig)
}

type SharedOption interface {
	Option
	CaptureOption
}

type reporterOption struct {
	reporter progress.Reporter
}

func (o *reporterOption) applyArchiver(c *archiverConfig) {
	c.reporter = o.reporter
}

func (o *reporterOption) applyCapture(c *captureConfig) {
	c.reporter = o.reporter
}

func WithReporter(reporter progress.Reporter) SharedOption {
	return &reporterOption{
		reporter: reporter,
	}
}

type userAgentOption struct {
	userAgent string
}

func (o *userAgentOption) applyArchiver(c *archiverConfig) {
	c.userAgent = o.userAgent
}

func (o *userAgentOption) applyCapture(c *captureConfig) {
	c.userAgent = o.userAgent
}

func WithUserAgent(ua string) SharedOption {
	return &userAgentOption{
		userAgent: ua,
	}
}

type screenshotOption struct {
	f func([]byte) error
}

func (o *screenshotOption) applyCapture(c *captureConfig) {
	c.screenshotFunc = o.f
}

func WithScreenshot(screenshotFunc func([]byte) error) CaptureOption {
	return &screenshotOption{
		f: screenshotFunc,
	}
}
