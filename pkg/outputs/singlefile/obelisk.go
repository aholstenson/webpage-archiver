package singlefile

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base32"
	"mime"
	"net/http"
	"net/http/httputil"
	"os"
	"path"

	"github.com/aholstenson/webpage-archiver/pkg/outputs"
	"github.com/go-shiori/obelisk"
	"github.com/rosshhun/gonormalizer"
)

type SingleFileOutput struct {
	tmpDir   string
	filename string
	url      string
}

func NewOutput(filename string) (*SingleFileOutput, error) {
	tmpDir, err := os.MkdirTemp("", "webpage-archiver")
	if err != nil {
		return nil, err
	}

	return &SingleFileOutput{
		tmpDir:   tmpDir,
		filename: filename,
	}, nil
}

func (o *SingleFileOutput) Close() error {
	defer os.RemoveAll(o.tmpDir)

	archiver := obelisk.Archiver{
		Transport: o,
	}
	archiver.Validate()

	data, ct, err := archiver.Archive(context.Background(), obelisk.Request{
		URL: o.url,
	})
	if err != nil {
		return err
	}

	ext := ".bin"
	extensions, err := mime.ExtensionsByType(ct)
	if err == nil && extensions != nil && len(extensions) > 0 {
		// Loop through and pick out the longest extension
		ext = extensions[0]
		for _, e := range extensions[1:] {
			if len(e) > len(ext) {
				ext = e
			}
		}
	}

	return os.WriteFile(o.filename+ext, data, 0666)
}

func (o *SingleFileOutput) Request(req *http.Request) error {
	if o.url == "" {
		o.url = req.URL.String()
	}
	return nil
}

func (o *SingleFileOutput) Response(req *http.Request, res *http.Response) error {
	data, err := httputil.DumpResponse(res, true)
	if err != nil {
		return err
	}
	path := o.pathTo(req.URL.String())
	return os.WriteFile(path, data, 0666)
}

func (o *SingleFileOutput) RoundTrip(req *http.Request) (*http.Response, error) {
	path := o.pathTo(req.URL.String())
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// TODO: This response might not be enough for Obelisk, tests needed
		return &http.Response{
			StatusCode: 404,
			Request:    req,
		}, nil
	} else if err != nil {
		return nil, err
	}

	res, err := http.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 300 && res.StatusCode < 400 {
		location := res.Header.Get("Location")
		if location != "" {
			req2, err := http.NewRequest("GET", location, nil)
			if err != nil {
				return nil, err
			}

			return o.RoundTrip(req2)
		}
	}

	return res, nil
}

func (o *SingleFileOutput) pathTo(url string) string {
	nurl, err := gonormalizer.Normalize(url)
	if err != nil {
		// If there's an error ignore it and try with the original URL
		nurl = url
	}

	hash := sha256.New()
	hash.Write([]byte(nurl))
	id := base32.HexEncoding.EncodeToString(hash.Sum(make([]byte, 0)))
	return path.Join(o.tmpDir, id)
}

var _ outputs.Output = &SingleFileOutput{}
