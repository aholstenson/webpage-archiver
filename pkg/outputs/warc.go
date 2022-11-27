package outputs

import (
	"strconv"
	"time"

	"github.com/aholstenson/webpage-archiver/pkg/network"
	"github.com/nlnwa/gowarc"
)

type WARCOutput struct {
	writer *gowarc.WarcFileWriter
}

func NewWARCOutput(directory string) (*WARCOutput, error) {
	writer := gowarc.NewWarcFileWriter(gowarc.WithFileNameGenerator(&gowarc.PatternNameGenerator{
		Directory: directory,
	}))

	return &WARCOutput{
		writer: writer,
	}, nil
}

func (o *WARCOutput) Close() error {
	return o.writer.Close()
}

func (o *WARCOutput) Request(req *network.Request) error {
	return nil
}

func (o *WARCOutput) Response(res *network.Response) error {
	builder := gowarc.NewRecordBuilder(gowarc.Response)

	_, err := builder.WriteString("HTTP/1.1 " + strconv.Itoa(res.StatusCode) + " " + res.StatusPhrase + "\n")
	if err != nil {
		return err
	}

	for key, values := range res.Headers {
		for _, value := range values {
			_, err = builder.WriteString(key + ": " + value + "\n")
			if err != nil {
				return err
			}
		}
	}

	_, err = builder.WriteString("\n")
	if err != nil {
		return err
	}

	_, err = builder.Write(res.Body)
	if err != nil {
		return err
	}

	builder.AddWarcHeader(gowarc.WarcTargetURI, res.URL)
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
