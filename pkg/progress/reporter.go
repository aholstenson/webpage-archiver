package progress

import (
	"io"

	"github.com/aholstenson/webpage-archiver/pkg/network"
)

type Reporter interface {
	io.Closer

	Start(url string)

	Debug(msg string)

	Error(err error, msg string)

	Request(req *network.Request)

	Response(res *network.Response)
}
