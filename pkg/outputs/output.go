package outputs

import (
	"io"
	"net/http"
)

type Output interface {
	io.Closer

	Request(req *http.Request) error

	Response(req *http.Request, res *http.Response) error
}
