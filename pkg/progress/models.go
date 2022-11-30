package progress

// Request represents basic information about a request.
type Request struct {
	// URL being requested.
	URL string
	// Method used to access the request.
	Method string
}

// Response contains information about a response.
type Response struct {
	// URL of the response.
	URL string
	// StatusCode is the HTTP status code of the request, such as 200, 301,
	// 404 etc.
	StatusCode int
	// StatusPhrase is the phrase associated with the HTTP status code.
	StatusPhrase string
	// BodySize is the number of bytes of the body.
	BodySize int
}
