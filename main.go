package main

import (
	"context"
	"os"
	"path"
	"time"

	"github.com/aholstenson/webpage-archiver/pkg/captures"
	"github.com/aholstenson/webpage-archiver/pkg/outputs"
	"github.com/aholstenson/webpage-archiver/pkg/progress"
	"github.com/alecthomas/kong"
	"github.com/mattn/go-isatty"
)

func main() {
	var cli struct {
		WARC       string `type:"existingdir" required:"" group:"warc" xor:"singlefile,warc" help:"Directory where WARC files will be stored"`
		SingleFile string `type:"existingdir" required:"" group:"singlefile" xor:"singlefile,warc" help:"Directory where HTML files will be stored"`

		Screenshot bool `help:"Enable screenshots alongside other stored files"`

		URL []string `arg:"" required:"" help:"URLs to capture"`
	}

	cliCtx := kong.Parse(&cli, kong.UsageOnError())
	cliCtx.FatalIfErrorf(cliCtx.Error)

	prefix := time.Now().In(time.UTC).Format("20060102150405") + "-"

	var err error
	var directory string
	var output outputs.Output
	if cli.WARC != "" {
		directory = cli.WARC
		output, err = outputs.NewWARCOutput(cli.WARC, prefix)
		if err != nil {
			cliCtx.Errorf("Could not create WARC output: %s", err.Error())
			return
		}
	} else if cli.SingleFile != "" {
		directory = cli.SingleFile
		output, err = outputs.NewObeliskOutput(path.Join(directory, prefix))
		if err != nil {
			cliCtx.Errorf("Could not create single file output: %s", err.Error())
			return
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var reporter progress.Reporter
	if isatty.IsTerminal(os.Stdout.Fd()) {
		reporter, err = progress.NewInteractiveReporter(func() {
			reporter.Info("Exiting")
			cancel()
		})
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

	options := []captures.Option{
		captures.WithReporter(reporter),
		captures.WithOutput(output),
	}
	if cli.Screenshot {
		options = append(
			options,
			captures.WithScreenshots(directory, prefix),
		)
	}

	capturer, err := captures.NewCapturer(options...)
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

	reporter.Info("Finalizing output")
	output.Close()
	reporter.Info("Closing browser")
	capturer.Close()
	reporter.Close()
}
