package outputs

import (
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/nlnwa/gowarc"
)

type WARCOutput struct {
	writer *gowarc.WarcFileWriter
}

func NewWARCOutput(directory string, prefix string) (*WARCOutput, error) {
	writer := gowarc.NewWarcFileWriter(gowarc.WithFileNameGenerator(&gowarc.PatternNameGenerator{
		Directory: directory,
		Pattern:   prefix + "%04{serial}d.%{ext}s",
	}))

	return &WARCOutput{
		writer: writer,
	}, nil
}

func (o *WARCOutput) Close() error {
	return o.writer.Close()
}

func (o *WARCOutput) Request(req *http.Request) error {
	return nil
}

func (o *WARCOutput) Response(req *http.Request, res *http.Response) error {
	builder := gowarc.NewRecordBuilder(gowarc.Response)

	data, err := httputil.DumpResponse(res, true)
	if err != nil {
		return err
	}

	_, err = builder.Write(data)
	if err != nil {
		return err
	}

	builder.AddWarcHeader(gowarc.WarcTargetURI, req.URL.String())
	builder.AddWarcHeaderTime(gowarc.WarcDate, time.Now())
	builder.AddWarcHeader(gowarc.ContentType, "application/http; msgtype=response")

	record, _, err := builder.Build()
	if err != nil {
		return err
	}

	o.writer.Write(record)
	return nil
}

var _ Output = &WARCOutput{}
