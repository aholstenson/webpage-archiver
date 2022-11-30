# webpage-archiver

Capture and archive webpages to WARC-files, available both as a command-line
tool and as a Go library.

Features:

- Customizable output
  - WARC
  - Single file support via [Obelisk](https://github.com/go-shiori/obelisk)
- Screenshot support
- Archives using a headless Chrome instance
  - Will automatically download a compatible headless browser

## Capturing pages

Store WARC files in a specific directory:

```console
webpage-archiver --output directory/ urlToArchive
```

To archive as a single file instead:

```console
webpage-archiver --output fileOrDirectory --single-file urlToArchive
```

Storing a screenshot of each page can be done with `--screenshot`:

```console
webpage-archiver --output directory/ --screenshot urlToArchive
```

Multiple URLs can be captured to the same archive:

```console
webpage-archiver --output directory/ urlToArchive anotherUrlToArchive
```

## Viewing pages

WARC-files captured with this tool need to be replayed, the easiest way to
replay a capture is to use a tool like [ReplayWeb.page](https://replayweb.page/).

## Using as Go Library

```console
go get github.com/aholstenson/webpage-archiver
```

Create an archiver instance to start capturing pages:

```go
archiver, err := archiver.NewArchiver(ctx)
```

Capture pages using `Capture`:

```go
output, err := warc.NewOutput(warc.WithDirectory(directory))

err := archiver.Capture(ctx, url, output)

output.Close()
```

Close the archiver when it's no longer needed:

```go
archiver.Close()
```

### Tracking progress

Archiver can take an optional progress reporter that will be used to log
actions, requests and responses:

```go
reporter := progress.NewConsoleReporter()
archiver := archiver.NewArchiver(ctx, archiver.WithProgress(reporter))
```

To use a progress reporter for a specific capture:

```go
err := archiver.Capture(ctx, url, output, archiver.WithProgress(reporter))
```

### User agents

The user agent can be specified via `WithUserAgent`:

```go
archiver.WithUserAgent("Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible) Safari/537.36")
```

This option can be applied both to `NewArchiver` and to `Archiver.Capture`.

### Screenshots

The option `WithScreenshot` can be passed to `Capture` to receive a screenshot
of the page as it looks before the archiving ends.

```go
archiver.Capture(ctx, url, output, archiver.WithScreenshot(func(data []byte) error {
  // Handle screenshot data here
  return nil
}))
```
