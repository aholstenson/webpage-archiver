package outputs

import (
	"io"
	"net/http"
)

type Output interface {
	io.Closer

	StartPage(url string) error

	FinishPage(url string) error

	Request(req *http.Request) error

	Response(req *http.Request, res *http.Response) error
}
