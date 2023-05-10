package io

import (
	"fmt"
)

type ioStringStreamImpl struct {
	index   int
	input   string
	outputs *[]string
}

func NewStringStream(input string, outputs *[]string) Io {
	return &ioStringStreamImpl{
		index:   len(*outputs) - 1,
		input:   input,
		outputs: outputs,
	}
}

func (self *ioStringStreamImpl) Scan(
	format string,
	args ...interface{},
) (int, error) {
	return fmt.Sscanf(self.input, format, args)
}

func (self *ioStringStreamImpl) Print(msg string, data ...interface{}) error {
	if self.index+1 >= len(*(self.outputs)) {
		self.index = len(*(self.outputs)) - 1
	}

	if self.index < 0 {
		*(self.outputs) = append(*(self.outputs), "")
		self.index++
	}

	(*self.outputs)[self.index] += fmt.Sprintf(msg, data...)
	return nil
}

func (self *ioStringStreamImpl) Write(b []byte) (int, error) {
	if self.index+1 >= len(*(self.outputs)) {
		self.index = len(*(self.outputs)) - 1
	}

	if self.index < 0 {
		*(self.outputs) = append(*(self.outputs), "")
		self.index++
	}

	(*self.outputs)[self.index] += string(b)
	return len(b), nil
}

func (self *ioStringStreamImpl) Flush() {
	if self.index+1 >= len(*(self.outputs)) {
		self.index = len(*(self.outputs)) - 1
	}

	*(self.outputs) = append(*(self.outputs), "")
	self.index++
}
