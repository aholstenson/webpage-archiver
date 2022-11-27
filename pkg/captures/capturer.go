package captures

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aholstenson/webpage-archiver/pkg/network"
	"github.com/aholstenson/webpage-archiver/pkg/outputs"
	"github.com/aholstenson/webpage-archiver/pkg/progress"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/go-rod/stealth"
)

type Capturer struct {
	reporter progress.Reporter
	output   outputs.Output

	browser    *rod.Browser
	httpClient *http.Client

	userAgent string
}

func NewCapturer(opts ...Option) (*Capturer, error) {
	options := &capturerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	reporter := options.reporter
	var log utils.Log = func(msg ...any) {
		reporter.Debug(fmt.Sprint(msg...))
	}

	browser := rod.New().Logger(log)
	err := browser.Connect()
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Capturer{
		reporter: options.reporter,
		output:   options.output,

		browser: browser,

		httpClient: httpClient,
		userAgent:  options.userAgent,
	}, nil
}

func (c *Capturer) Close() error {
	return c.browser.Close()
}

func (c *Capturer) Capture(ctx context.Context, requestURL string) {
	page, err := stealth.Page(c.browser)
	if err != nil {
		c.reporter.Error(err, "Could not fetch webpage")
		return
	}

	page = page.Context(ctx)

	page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:  1920,
		Height: 1080,
	})

	if c.userAgent != "" {
		page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
			UserAgent: c.userAgent,
		})
	}

	router := page.HijackRequests()
	router.MustAdd("", func(ctx *rod.Hijack) {
		request := &network.Request{
			URL:     ctx.Request.URL().String(),
			Method:  ctx.Request.Method(),
			Headers: network.Header(ctx.Request.Req().Header),
		}
		err := c.output.Request(request)
		if err != nil {
			c.reporter.Error(err, "Could not write request")
			return
		}

		c.reporter.Request(request)

		err = ctx.LoadResponse(c.httpClient, true)
		if err != nil {
			c.reporter.Error(err, "Could not load response")
			return
		}

		response := &network.Response{
			URL:          ctx.Request.URL().String(),
			Headers:      network.Header(ctx.Response.Headers()),
			StatusCode:   ctx.Response.Payload().ResponseCode,
			StatusPhrase: ctx.Response.Payload().ResponsePhrase,
			Body:         ctx.Response.Payload().Body,
		}
		if response.StatusPhrase == "" {
			response.StatusPhrase = http.StatusText(response.StatusCode)
		}

		err = c.output.Response(response)
		if err != nil {
			c.reporter.Error(err, "Could write response")
			return
		}

		c.reporter.Response(response)
	})
	go router.Run()

	err = page.Navigate(requestURL)
	if err != nil {
		c.reporter.Error(err, "Could not navigate to URL")
		router.Stop()
		return
	}

	// Wait for the page to be considered loaded
	page.MustWaitLoad()

	// Wait for the page to be idle when it comes to network traffic
	waiter := page.WaitRequestIdle(1*time.Second, nil, nil)
	waiter()

	// Stop the router
	router.Stop()
	page.Close()
}