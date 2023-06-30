package common

import "io"

type IObjectData interface {
	Bytes() []byte
	Reader() io.ReadCloser
}

type ObjectData struct {
	reader io.ReadCloser
	// size   int64
}

func NewObjectData(reader io.ReadCloser) *ObjectData {
	return &ObjectData{
		reader: reader,
	}
}

func (d *ObjectData) Bytes() []byte {
	data, _ := io.ReadAll(d.reader)
	return data
}

func (d *ObjectData) Reader() io.ReadCloser {
	return d.reader
}

func (d *ObjectData) Close() error {
	return d.reader.Close()
}
