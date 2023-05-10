package main

import (
	"fmt"
	"testing"

	alg "github.com/hung0913208/go-algorithm/lib/algorithm"
)

var table = []struct {
	input int
}{
	{input: 100},
	{input: 1000},
	{input: 74382},
	{input: 382399},
}

func BenchmarkHeap(b *testing.B) {
	heap := alg.NewIntHeap(true)

	for _, v := range table {
		b.Run(
			fmt.Sprintf("input_size_%d", v.input),
			func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					heap.Push(i)
				}

				for i := 0; i < b.N; i++ {
					heap.Get()
					heap.Pop()
				}
			},
		)
	}
}
