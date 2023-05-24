package heap

import (
	"errors"
)

type Comparator func(l, r interface{}) int

type Heap interface {
	Push(value interface{}) error
	Pop() error
	Get() (interface{}, error)
	Size() int
}

type heapImpl struct {
	branch, all int
	array       []interface{}
	compare     Comparator
}

func NewIntHeap(isAsc bool) Heap {
	return NewHeapWithComparator(
		func(l, r interface{}) int {
			lval := l.(int)
			rval := r.(int)

			if isAsc {
				if lval < rval {
					return 1
				} else if lval == rval {
					return 0
				} else {
					return -1
				}
			} else {
				if lval > rval {
					return 1
				} else if lval == rval {
					return 0
				} else {
					return -1
				}
			}
		},
	)
}

func NewHeapWithComparator(compare Comparator) Heap {
	return &heapImpl{
		array:   make([]interface{}, 0),
		branch:  0,
		all:     0,
		compare: compare,
	}
}

func (self *heapImpl) Push(value interface{}) error {
	self.array = append(self.array, value)
	self.all = self.all + 1
	self.branch = self.all / 2

	if !self.refresh(false) {
		return errors.New("can't append new value to heap")
	}

	return nil
}

func (self *heapImpl) Pop() error {
	if !self.refresh(true) {
		return errors.New("can't pop out the largest value out of heap")
	}

	return nil
}

func (self *heapImpl) Get() (interface{}, error) {
	if self.all == 0 || len(self.array) == 0 {
		return nil, errors.New("the heap is empty")
	}

	return self.array[0], nil
}

func (self *heapImpl) Size() int {
	return self.all
}

func (self *heapImpl) refresh(sorting bool) bool {
	if self.branch > 0 && sorting {
		return false
	}

	for {
		var copyBranch, leftLeaf, rightLeaf, cache int

		if self.branch > 0 {
			self.branch--
		} else if sorting {
			self.all--

			if self.all > 0 {
				self.swap(0, self.all)
			}
		} else {
			break
		}

		copyBranch = self.branch

		for {
			leftLeaf = 2*copyBranch + 1
			rightLeaf = leftLeaf + 1

			if rightLeaf < self.all {
				isSmaller := self.compare(
					self.array[leftLeaf],
					self.array[rightLeaf],
				)

				if isSmaller > 0 {
					copyBranch = leftLeaf
				} else {
					copyBranch = rightLeaf
				}
			} else {
				break
			}
		}

		if rightLeaf == self.all {
			copyBranch = leftLeaf
		}

		for {
			if copyBranch == self.branch {
				break
			}
			if self.compare(self.array[self.branch], self.array[copyBranch]) <= 0 {
				break
			}

			copyBranch = (copyBranch+copyBranch%2)/2 - 1
		}

		cache = copyBranch
		for copyBranch != self.branch {
			copyBranch = (copyBranch+copyBranch%2)/2 - 1
			self.swap(copyBranch, cache)
		}

		if self.branch == 0 && sorting {
			break
		}
	}

	return true
}

func (self *heapImpl) swap(l, r int) {
	self.array[l], self.array[r] = self.array[r], self.array[l]
}
