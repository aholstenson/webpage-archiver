package outputs

import (
	"io"

	"github.com/aholstenson/webpage-archiver/pkg/network"
)

type Output interface {
	io.Closer

	Request(req *network.Request) error

	Response(req *network.Response) error
}
