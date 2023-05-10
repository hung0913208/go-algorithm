package io

import (
	"io"
)

type Writer interface {
    Write(points []interface{}) error
	Flush()
}

type Io interface {
	io.Writer

	Scan(format string, args ...interface{}) (int, error)
	Print(msg string, data ...interface{}) error
	Flush()
}
