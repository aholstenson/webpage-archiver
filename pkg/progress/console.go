package progress

import (
	"os"
	"strconv"
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

func (c *consoleReporter) Action(msg string) {
	c.print("🚀 " + msg)
}

func (c *consoleReporter) Info(msg string) {
	c.print(msg)
}

func (c *consoleReporter) Debug(msg string) {
	c.print(msg)
}

func (c *consoleReporter) Error(err error, msg string) {
	c.print("❌ " + msg + ": " + err.Error())
}

func (c *consoleReporter) Request(req *Request) {
	c.print("⬆️ " + req.Method + " " + req.URL)
}

func (c *consoleReporter) Response(res *Response) {
	c.print("⬇️ " + strconv.Itoa(res.StatusCode) + " " + res.URL)
}

var _ Reporter = &consoleReporter{}
