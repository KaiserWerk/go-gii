package gii

import (
	"archive/zip"
	"bytes"
	"io"
)

func Unzip(source io.Reader) ([]*zip.File, error) {
	data, err := io.ReadAll(source)
	if err != nil {
		return nil, err
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	return r.File, nil
}
