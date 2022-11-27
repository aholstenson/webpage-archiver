package main

import (
	"context"
	"os"
	"time"

	"github.com/aholstenson/webpage-archiver/pkg/captures"
	"github.com/aholstenson/webpage-archiver/pkg/outputs"
	"github.com/aholstenson/webpage-archiver/pkg/progress"
	"github.com/alecthomas/kong"
	"github.com/mattn/go-isatty"
)

func main() {
	var cli struct {
		WARC string `type:"existingdir" required:"" help:"Directory where WARC files will be stored"`

		URL []string `arg:"" required:"" help:"URLs to capture"`
	}

	cliCtx := kong.Parse(&cli, kong.UsageOnError())
	cliCtx.FatalIfErrorf(cliCtx.Error)

	warcOutput, err := outputs.NewWARCOutput(cli.WARC)
	if err != nil {
		cliCtx.Errorf("Could not create WARC output: %s", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var reporter progress.Reporter
	if isatty.IsTerminal(os.Stdout.Fd()) {
		reporter, err = progress.NewInteractiveReporter(cancel)
		if err != nil {
			cliCtx.Errorf("Could not create console reporter: %s", err.Error())
			return
		}
	} else {
		reporter, err = progress.NewConsoleReporter()
		if err != nil {
			cliCtx.Errorf("Could not create console reporter: %s", err.Error())
			return
		}
	}

	capturer, err := captures.NewCapturer(
		captures.WithReporter(reporter),
		captures.WithOutput(warcOutput),
	)
	if err != nil {
		cliCtx.Errorf("Could not capture: %s", err.Error())
		return
	}

	for _, url := range cli.URL {
		func() {
			ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			capturer.Capture(ctx, url)
		}()
	}

	warcOutput.Close()
	reporter.Close()
	capturer.Close()
}
