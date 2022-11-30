package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/aholstenson/webpage-archiver/pkg/archiver"
	"github.com/aholstenson/webpage-archiver/pkg/outputs"
	"github.com/aholstenson/webpage-archiver/pkg/outputs/singlefile"
	"github.com/aholstenson/webpage-archiver/pkg/outputs/warc"
	"github.com/aholstenson/webpage-archiver/pkg/progress"
	"github.com/alecthomas/kong"
	"github.com/mattn/go-isatty"
)

type CLI struct {
	Output string `type="path" short:"o" help:"Output directory or file" default:"."`

	WARC       bool `group:"warc" xor:"singlefile,warc" help:"Store pages in WARC files"`
	SingleFile bool `group:"singlefile" xor:"singlefile,warc" help:"Store pages as single-file HTML"`

	Screenshot bool `help:"Enable screenshots alongside other stored files"`

	URL []string `arg:"" required:"" help:"URLs to capture"`
}

func Run() {
	cli := &CLI{}
	cliCtx := kong.Parse(cli, kong.UsageOnError())
	cliCtx.FatalIfErrorf(cliCtx.Error)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	var reporter progress.Reporter
	if isatty.IsTerminal(os.Stdout.Fd()) {
		reporter, err = newInteractiveReporter(func() {
			reporter.Info("Exiting")
			cancel()
		})
		if err != nil {
			cliCtx.Fatalf("Could not create console reporter: %s", err.Error())
			return
		}
	} else {
		reporter, err = progress.NewConsoleReporter()
		if err != nil {
			cliCtx.Fatalf("Could not create console reporter: %s", err.Error())
			return
		}
	}

	err = cli.run(ctx, cancel, reporter)
	cliCtx.FatalIfErrorf(err)
}

func (cli *CLI) run(
	ctx context.Context,
	exitFunc func(),
	reporter progress.Reporter,
) error {
	// TODO: Support for custom prefixes
	prefix := time.Now().In(time.UTC).Format("20060102150405") + "-"

	var directory string
	var outputFactory Outputs
	if cli.SingleFile {
		isDir, err := IsDir(cli.Output)
		if err != nil {
			return fmt.Errorf("could not check if %q is a directory: %w", cli.Output, err)
		}

		if !isDir {
			directory = path.Dir(cli.Output)
			isParentDir, err := IsDir(directory)
			if err != nil {
				return fmt.Errorf("could not check if %q is a directory: %w", directory, err)
			} else if !isParentDir {
				return fmt.Errorf("%q must be an existing directory", directory)
			}

			filename := path.Base(cli.Output)
			ext := path.Ext(filename)
			prefix = filename[0 : len(filename)-len(ext)]

			outputFactory = &MultiOutput{
				Create: func(seq int64) (outputs.Output, error) {
					return singlefile.NewOutput(cli.Output)
				},
			}
		} else {
			directory = cli.Output
			outputFactory = &MultiOutput{
				Create: func(seq int64) (outputs.Output, error) {
					return singlefile.NewOutput(path.Join(directory, fmt.Sprintf("%s%04d", prefix, seq)))
				},
			}
		}
	} else {
		isDir, err := IsDir(cli.Output)
		if err != nil {
			return fmt.Errorf("could not check if %q is a directory: %w", cli.Output, err)
		} else if !isDir {
			return fmt.Errorf("%q is not a directory", cli.Output)
		}

		directory = cli.Output
		output, err := warc.NewOutput(directory, warc.WithPrefix(prefix))
		if err != nil {
			return fmt.Errorf("could not create WARC output: %w", err)
		}

		outputFactory = &SingleOutput{Output: output}
	}

	capturer, err := archiver.NewArchiver(archiver.WithReporter(reporter))
	if err != nil {
		return fmt.Errorf("could not create archiver: %w", err)
	}

	seq := 0
	for _, url := range cli.URL {
		if ctx.Err() != nil {
			break
		}

		seq++
		func() {
			options := []archiver.CaptureOption{}
			if cli.Screenshot {
				options = append(
					options,
					archiver.WithScreenshot(func(b []byte) error {
						return os.WriteFile(
							path.Join(directory, fmt.Sprintf("%s%04d.png", prefix, seq)),
							b,
							0644)
					}),
				)
			}

			output, err := outputFactory.Get(url)
			if err != nil {
				reporter.Error(err, "Failed to create output")
				exitFunc()
				return
			}

			ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			capturer.Capture(ctx, url, output, options...)

			err = output.Close()
			if err != nil {
				reporter.Error(err, "Failed to write output")
				exitFunc()
				return
			}
		}()
	}

	reporter.Info("Finalizing output")
	outputFactory.Close()
	reporter.Info("Closing browser")
	capturer.Close()

	if closer, ok := reporter.(io.Closer); ok {
		closer.Close()
	}
	return nil
}
