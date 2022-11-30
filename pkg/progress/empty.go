package progress

type emptyReporter struct {
}

func NewEmptyReporter() Reporter {
	return &emptyReporter{}
}

func (c *emptyReporter) Action(msg string) {
}

func (c *emptyReporter) Info(msg string) {
}

func (c *emptyReporter) Debug(msg string) {
}

func (c *emptyReporter) Error(err error, msg string) {
}

func (c *emptyReporter) Request(req *Request) {
}

func (c *emptyReporter) Response(res *Response) {
}

var _ Reporter = &emptyReporter{}
