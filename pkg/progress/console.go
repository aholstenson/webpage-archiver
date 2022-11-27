package progress

import (
	"os"
	"strconv"

	"github.com/aholstenson/webpage-archiver/pkg/network"
)

type consoleReporter struct {
}

func NewConsoleReporter() (Reporter, error) {
	return &consoleReporter{}, nil
}

func (c *consoleReporter) print(msg string) {
	os.Stdout.Write([]byte(msg + "\n"))
}

func (c *consoleReporter) Close() error {
	return nil
}

func (c *consoleReporter) Start(url string) {
	c.print("🌎 " + url)
}

func (c *consoleReporter) Debug(msg string) {
	c.print(msg)
}

func (c *consoleReporter) Error(err error, msg string) {
	c.print("❌ " + msg + ": " + err.Error())
}

func (c *consoleReporter) Request(req *network.Request) {
	c.print("⬆️ " + req.Method + " " + req.URL)
}

func (c *consoleReporter) Response(res *network.Response) {
	c.print("⬇️ " + strconv.Itoa(res.StatusCode) + " " + res.URL)
}

var _ Reporter = &consoleReporter{}
