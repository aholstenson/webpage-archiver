package network

import "net/http"

type Header http.Header

type Request struct {
	URL     string
	Method  string
	Headers Header
}

type Response struct {
	URL          string
	StatusCode   int
	StatusPhrase string
	Headers      Header
	Body         []byte
}
