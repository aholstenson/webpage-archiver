package main

import (
	"context"
	"time"

	"github.com/aholstenson/webpage-archiver/pkg/captures"
	"github.com/aholstenson/webpage-archiver/pkg/outputs"
	"github.com/aholstenson/webpage-archiver/pkg/progress"
	"github.com/alecthomas/kong"
)

func main() {
	var cli struct {
		WARC string `type:"existingdir" required:"" help:"Directory where WARC files will be stored"`

		URL []string `arg:"" required:"" help:"URLs to capture"`
	}

	ctx := kong.Parse(&cli, kong.UsageOnError())
	ctx.FatalIfErrorf(ctx.Error)

	warcOutput, err := outputs.NewWARCOutput(cli.WARC)
	if err != nil {
		ctx.Errorf("Could not create WARC output: %s", err.Error())
		return
	}

	reporter := &progress.ConsoleReporter{}
	capturer, err := captures.NewCapturer(
		captures.WithReporter(reporter),
		captures.WithOutput(warcOutput),
	)
	if err != nil {
		ctx.Errorf("Could not capture: %s", err.Error())
		return
	}

	for _, url := range cli.URL {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
			defer cancel()
			capturer.Capture(ctx, url)
		}()
	}

	warcOutput.Close()
	capturer.Close()
}
