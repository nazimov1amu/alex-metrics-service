package encoding

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
)

type CompressorResponseWriter struct {
	rw   http.ResponseWriter
	grw *gzip.Writer
}

func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decompress(data io.ReadCloser) (io.ReadCloser, error) {
	gz, err := gzip.NewReader(data)
	if err != nil {
		return nil, err
	}
	return gz, nil
}

