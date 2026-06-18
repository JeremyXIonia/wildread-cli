package parser

import (
	"bytes"
	"io"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
)

func newBytesReader(b []byte) io.Reader { return bytes.NewReader(b) }

func readAll(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	return string(b), err
}

func lookupEncoding(label string) (encoding.Encoding, error) {
	return htmlindex.Get(label)
}
