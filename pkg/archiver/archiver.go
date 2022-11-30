package archiver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/aholstenson/webpage-archiver/pkg/outputs"
	"github.com/aholstenson/webpage-archiver/pkg/progress"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/go-rod/stealth"
)

type Archiver struct {
	reporter  progress.Reporter
	userAgent string

	browser    *rod.Browser
	httpClient *http.Client
	request    atomic.Int32
}

func NewArchiver(opts ...Option) (*Archiver, error) {
	config := &archiverConfig{}
	for _, opt := range opts {
		opt.applyArchiver(config)
	}

	reporter := config.reporter
	var log utils.Log = func(msg ...any) {
		reporter.Debug(fmt.Sprint(msg...))
	}

	reporter.Action("Finding browser")
	browserDownloader := launcher.NewBrowser()
	browserDownloader.Logger = log
	browserBin, err := browserDownloader.Get()
	if err != nil {
		return nil, err
	}

	reporter.Action("Starting browser")
	launcher := launcher.New().Bin(browserBin)
	controlURL, err := launcher.Launch()
	if err != nil {
		return nil, err
	}

	browser := rod.New().ControlURL(controlURL).Logger(log)
	err = browser.Connect()
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Archiver{
		reporter: config.reporter,

		browser: browser,

		httpClient: httpClient,
		userAgent:  config.userAgent,
	}, nil
}

func (c *Archiver) Close() error {
	return c.browser.Close()
}

func (c *Archiver) Capture(
	ctx context.Context,
	requestURL string,
	output outputs.Output,
	opts ...CaptureOption,
) {
	config := &captureConfig{
		reporter:  c.reporter,
		userAgent: c.userAgent,
	}
	for _, opt := range opts {
		opt.applyCapture(config)
	}

	reporter := config.reporter
	reporter.Action(requestURL)

	page, err := stealth.Page(c.browser)
	if err != nil {
		reporter.Error(err, "Could not fetch webpage")
		return
	}

	page = page.Context(ctx)

	page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:  1920,
		Height: 1080,
	})

	if config.userAgent != "" {
		page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
			UserAgent: c.userAgent,
		})
	}

	router := page.HijackRequests()
	err = router.Add("", "", func(ctx *rod.Hijack) {
		request := &progress.Request{
			URL:    ctx.Request.URL().String(),
			Method: ctx.Request.Method(),
		}
		err := output.Request(ctx.Request.Req())
		if err != nil {
			reporter.Error(err, "Could not write request")
			return
		}

		reporter.Request(request)

		res, err := c.httpClient.Do(ctx.Request.Req())
		if err != nil {
			var dnsError *net.DNSError
			if errors.As(err, &dnsError) {
				ctx.Response.Fail(proto.NetworkErrorReasonAddressUnreachable)
				reporter.Error(err, "Could not load response")
			} else if !errors.Is(err, context.Canceled) {
				ctx.Response.Fail(proto.NetworkErrorReasonConnectionFailed)
				reporter.Error(err, "Could not load response")
			} else {
				ctx.Response.Fail(proto.NetworkErrorReasonConnectionAborted)
			}
			return
		}

		defer func() { _ = res.Body.Close() }()

		ctx.Response.Payload().ResponseCode = res.StatusCode

		for k, vs := range res.Header {
			for _, v := range vs {
				ctx.Response.SetHeader(k, v)
			}
		}

		b, err := io.ReadAll(res.Body)
		if err != nil {
			ctx.Response.Fail(proto.NetworkErrorReasonConnectionAborted)
			return
		}
		ctx.Response.Payload().Body = b
		res.Body = io.NopCloser(bytes.NewBuffer(b))

		response := &progress.Response{
			URL:          ctx.Request.URL().String(),
			StatusCode:   ctx.Response.Payload().ResponseCode,
			StatusPhrase: ctx.Response.Payload().ResponsePhrase,
			BodySize:     len(ctx.Response.Payload().Body),
		}
		if response.StatusPhrase == "" {
			response.StatusPhrase = http.StatusText(response.StatusCode)
		}

		err = output.Response(ctx.Request.Req(), res)
		if err != nil {
			reporter.Error(err, "Could write response")
			return
		}

		reporter.Response(response)
	})
	if err != nil {
		reporter.Error(err, "Could not setup required request hijacking")
		return
	}
	go router.Run()

	defer router.Stop()
	defer page.Close()

	err = page.Navigate(requestURL)
	if err != nil {
		reporter.Error(err, "Could not navigate to URL")
		return
	}

	// Wait for the page to be considered loaded
	err = page.WaitLoad()
	if err != nil {
		reporter.Error(err, "Could not load page")
		return
	}

	idle := make(chan any)
	reporter.Info("Waiting for page to fully load")
	waiter := page.WaitRequestIdle(2*time.Second, nil, nil)
	go func() {
		// Wait for the page to be idle when it comes to network traffic
		waiter()
		idle <- struct{}{}
	}()

	// In an attempt to capture pages that lazy-load images and content we
	// implement a loop where we either wait for the context to be done or
	// the requests to be idle. While doing this we also scroll the site a
	// little bit to trigger loading of new content.
_outer:
	for {
		select {
		case <-ctx.Done():
			// The context has been canceled, return
			return
		case <-idle:
			// If network is idle stop waiting
			break _outer
		case <-time.After(100):
			// After a small delay we try to scroll a bit in an attempt to
			// capture the entire page.
			page.Mouse.Scroll(0, 50, 1)
		}
	}

	if config.screenshotFunc != nil {
		reporter.Info("Taking screenshot")

		// Update the viewport to scroll to the top
		page.AddScriptTag("", "window.scrollTo(0,0)")
		time.Sleep(100)

		data, err := page.Screenshot(false, &proto.PageCaptureScreenshot{
			Format: proto.PageCaptureScreenshotFormatPng,
		})
		if err != nil {
			reporter.Error(err, "Could not screenshot page")
			return
		}

		err = config.screenshotFunc(data)
		if err != nil {
			reporter.Error(err, "Could not handle screenshot")
			return
		}
	}
}
