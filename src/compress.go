package png

import (
	"bytes"
	"compress/zlib"
	"io"
)

type Compresser interface {
	Compress([]byte) ([]byte, error) // TODO: figure out signature
	Uncompress([]byte) ([]byte, error)
}

type Flater struct{}

func NewFlater() Compresser {
	return &Flater{}
}

func (*Flater) Compress(data []byte) ([]byte, error) {
	b := bytes.Buffer{}
	w := zlib.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (*Flater) Uncompress(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	b := bytes.Buffer{}
	io.Copy(&b, r)

	return b.Bytes(), nil
}

var _ Compresser = (*Flater)(nil)
