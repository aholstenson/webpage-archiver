package runner

import (
	"io"

	"github.com/aholstenson/webpage-archiver/pkg/outputs"
)

type NonCloseableOutput struct {
	outputs.Output
}

func (o *NonCloseableOutput) Close() error {
	return nil
}

type Outputs interface {
	io.Closer

	Get(url string) (outputs.Output, error)
}

type SingleOutput struct {
	Output outputs.Output
}

func (o *SingleOutput) Close() error {
	return o.Output.Close()
}

func (o *SingleOutput) Get(url string) (outputs.Output, error) {
	return &NonCloseableOutput{o.Output}, nil
}

type MultiOutput struct {
	Create func(seq int64) (outputs.Output, error)

	seq int64
}

func (o *MultiOutput) Close() error {
	return nil
}

func (o *MultiOutput) Get(url string) (outputs.Output, error) {
	o.seq++
	return o.Create(o.seq)
}
