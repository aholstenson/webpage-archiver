package progress

import (
	"io"

	"github.com/aholstenson/webpage-archiver/pkg/network"
)

type Reporter interface {
	io.Closer

	// Action sets the current action.
	Action(action string)

	// Info prints an information message to the log.
	Info(msg string)

	// Debug prints a debug message to the log.
	Debug(msg string)

	// Error prints an error message to the log.
	Error(err error, msg string)

	// Request is called when a request for a certain resource is started.
	Request(req *network.Request)

	// Response is called when a response for a request is received.
	Response(res *network.Response)
}
