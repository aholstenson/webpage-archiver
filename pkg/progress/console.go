package progress

import (
	"os"
	"strconv"

	"github.com/aholstenson/webpage-archiver/pkg/network"
)

type ConsoleReporter struct {
}

func (c *ConsoleReporter) print(msg string) {
	os.Stdout.Write([]byte(msg + "\n"))
}

func (c *ConsoleReporter) Close() error {
	return nil
}

func (c *ConsoleReporter) Debug(msg string) {
	c.print(msg)
}

func (c *ConsoleReporter) Error(err error, msg string) {
	c.print("❌ " + msg + ": " + err.Error())
}

func (c *ConsoleReporter) Request(req *network.Request) {
	c.print("⬆️ " + req.Method + " " + req.URL)
}

func (c *ConsoleReporter) Response(res *network.Response) {
	c.print("⬇️ " + strconv.Itoa(res.StatusCode) + " " + res.URL)
}

var _ Reporter = &ConsoleReporter{}
