package ahocorasick

import (
	"errors"
)

type AhoCorasick interface {
	Build(patterns ...string) error
	Optimize()
}

type vertexImpl struct {
	transition []int
	escape     int
	sufflink   int
}

type ahoCorasickImpl struct {
	array       []vertexImpl
	size        int
	dataRange   int
	isOptimized bool
	indexFunc   func(c rune, dataRange int) int
}

func NewAhoCorasick(
	dataRange int,
	indexFunc func(c rune, dataRange int) int,
) AhoCorasick {
	return NewAhoCorasickWithBuffer(dataRange, 0, indexFunc)
}

func NewAhoCorasickWithBuffer(
	dataRange int,
	bufferSize int,
	indexFunc func(c rune, dataRange int) int,
) AhoCorasick {
	return &ahoCorasickImpl{
		array:     make([]vertexImpl, bufferSize),
		size:      0,
		dataRange: dataRange,
		indexFunc: indexFunc,
	}
}

func (self *ahoCorasickImpl) Build(patterns ...string) error {
	if self.isOptimized {
		return errors.New("the aho-corasick object has been optimized")
	}

	for _, pattern := range patterns {
		vertex := 0

		for _, c := range pattern {
			leaf := self.indexFunc(c, self.dataRange)
			node := self.allocate(vertex)

			// @NOTE: assert that leaf should not be out of range
			if leaf >= len(node.transition) {
				panic("leaf out of range")
			}

			if node.transition[leaf] == 0 {
				node.transition[leaf] = self.size + 1
			}

			vertex = node.transition[leaf]
		}

		self.array[vertex].escape = vertex
	}

	return nil
}

func (self *ahoCorasickImpl) Optimize() {
}

func (self *ahoCorasickImpl) allocate(vertex int) *vertexImpl {
	if len(self.array) <= vertex {
		self.array = append(self.array, vertexImpl{
			transition: make([]int, self.dataRange),
			escape:     0,
			sufflink:   0,
		})
		self.size++
	} else if self.array[vertex].transition == nil {
		self.array[vertex].transition = make([]int, self.dataRange)
		self.size++
	}

	return &self.array[vertex]
}
