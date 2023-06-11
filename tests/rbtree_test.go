package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hung0913208/go-algorithm/lib/algorithm/rbtree"
)

func TestInsertRbTree(t *testing.T) {
	tree := rbtree.NewIntRbTree(true)

	for _, val := range []int{10, 5, 15, 3, 7, 20} {
		err := tree.Put(val, val)
		if err != nil {
			t.Errorf("put(%d) faces error: %v", val, err)
		}
	}

	rootKey, err := tree.Root()
	if err != nil {
		t.Errorf("get root faces error: %v", err)
	}
	if rootKey.(int) != 10 {
		t.Errorf("get wrong root key, actually we receive: %d", rootKey)
	}
	rootColor, err := tree.Color(rootKey)
	if err != nil {
		t.Errorf("color root faces error: %v", err)
	}
	if rootColor != rbtree.Black {
		t.Errorf("color root is not black, is actually %v", rootColor)
	}

	leftKey, err := tree.Left(rootKey)
	if err != nil {
		t.Errorf("get left faces error: %v", err)
	}
	if leftKey.(int) != 5 {
		t.Errorf("root's left must be 5 but we got: %d", leftKey)
	}
	leftColor, err := tree.Color(leftKey)
	if err != nil {
		t.Errorf("color root's left faces error: %v", err)
	}
	if leftColor != rbtree.Black {
		t.Errorf("color root's left is not black, is actually %v", leftColor)
	}

	rightKey, err := tree.Right(rootKey)
	if err != nil {
		t.Errorf("get right faces error: %v", err)
	}
	if rightKey.(int) != 15 {
		t.Errorf("root's right must be 15 but we got: %d", rightKey)
	}
	rightColor, err := tree.Color(rightKey)
	if err != nil {
		t.Errorf("color root's right faces error: %v", err)
	}
	if rightColor != rbtree.Black {
		t.Errorf("color root's right is not black, is actually %v", rightColor)
	}

	leftleftKey, err := tree.Left(leftKey)
	if err != nil {
		t.Errorf("get left left child faces error: %v", err)
	}
	if leftleftKey.(int) != 3 {
		t.Errorf("root's left left child must be 3 but we got: %d", leftleftKey)
	}
	leftleftColor, err := tree.Color(leftleftKey)
	if err != nil {
		t.Errorf("color root's left left child faces error: %v", err)
	}
	if leftleftColor != rbtree.Red {
		t.Errorf("color root's left left child is not red, is actually %v", leftleftColor)
	}

	leftrightKey, err := tree.Right(leftKey)
	if err != nil {
		t.Errorf("get left right child faces error: %v", err)
	}
	if leftrightKey.(int) != 7 {
		t.Errorf("root's left right child must be 7 but we got: %d", leftrightKey)
	}
	leftrightColor, err := tree.Color(leftrightKey)
	if err != nil {
		t.Errorf("color root's right right child faces error: %v", err)
	}
	if leftrightColor != rbtree.Red {
		t.Errorf("color root's left right child is not red, is actually %v", leftrightColor)
	}

	rightrightKey, err := tree.Right(rightKey)
	if err != nil {
		t.Errorf("get right right child faces error: %v", err)
	}
	if rightrightKey.(int) != 20 {
		t.Errorf("root's right right child must be 20 but we got: %d", rightrightKey)
	}
	rightrightColor, err := tree.Color(rightrightKey)
	if err != nil {
		t.Errorf("color root's right right child faces error: %v", err)
	}
	if rightrightColor != rbtree.Red {
		t.Errorf("color root's right right child is not red, is actually %v", rightrightColor)
	}
}

func TestInsertComplicatedToRbTree(t *testing.T) {
	tree := rbtree.NewIntRbTree(true)
	//samples := []int{50, 30, 70, 40, 10, 60, 80, 20}
	samples := []int{11, 2, 31, 34, 7, 35, 47, 0, 49, 26, 46, 3, 24, 13, 4, 37, 27, 12, 16, 10, 39, 19, 29, 25, 33, 41, 28, 45, 30, 48}

	for _, val := range samples {
		err := tree.Put(val, val)
		if err != nil {
			t.Errorf("put(%d) faces error: %v", val, err)
		}

		_, err = tree.Get(val)
		if err != nil {
			t.Errorf("insert(%d) not works", val)
			break
		}
	}

	err := tree.Iterate(func(key interface{}) error {
		left, _ := tree.Left(key)
		right, _ := tree.Right(key)

		color, err := tree.Color(key)
		if err != nil {
			tree.Lookup(
				key,
				true,
			)
			t.Errorf("get color fails (%d): %v", key, err)
			return err
		}

		if color == rbtree.Red {
			color, err = tree.Color(left)
			if err != nil {
				t.Errorf("get color fails: %v", err)
				return err
			}
			if color == rbtree.Red {
				t.Errorf("red's child (%d -> %d) must not be red", key, left)
			}

			color, err = tree.Color(right)
			if err != nil {
				t.Errorf("get color fails: %v", err)
				return err
			}
			if color == rbtree.Red {
				t.Errorf("red's child (%d -> %d) must not be red", key, right)
			}
		}

		return nil
	})
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestDeleteRbTree(t *testing.T) {
	tree := rbtree.NewIntRbTree(true)
	keys := []int{10, 5, 15, 3, 7, 12, 17, 1, 4, 6, 8, 11, 13, 16, 18, 2, 9}
	dels := []int{5, 15, 17, 10}

	// Create a Red-Black Tree with the following values
	for _, val := range keys {
		err := tree.Put(val, val)
		if err != nil {
			t.Errorf("put(%d) faces error: %v", val, err)
		}
	}

	// Delete the following values in the order given
	for _, val := range dels {
		err := tree.Delete(val)
		if err != nil {
			t.Errorf("delete(%d) fails: %v", val, err)
		}

		_, err = tree.Get(val)
		if err == nil {
			t.Errorf("delete(%d) not works", val)
		}
	}

	key, err := tree.Root()
	if err != nil {
		t.Errorf("get root's key fails: %v", err)
	}
	color, err := tree.Color(key)
	if err != nil {
		t.Errorf("color root's right faces error: %v", err)
	}
	if color != rbtree.Black {
		t.Errorf("color root is not black")
	}
	err = tree.Iterate(func(key interface{}) error {
		left, _ := tree.Left(key)
		right, _ := tree.Right(key)

		color, err := tree.Color(key)
		if err != nil {
			t.Errorf("get color fails (%d): %v", key, err)
			return err
		}

		if color == rbtree.Red {
			color, err = tree.Color(left)
			if err != nil {
				t.Errorf("get color fails: %v", err)
				return err
			}
			if color == rbtree.Red {
				t.Errorf("red's child (%d -> %d) must not be red", key, left)
			}

			color, err = tree.Color(right)
			if err != nil {
				t.Errorf("get color fails: %v", err)
				return err
			}
			if color == rbtree.Red {
				t.Errorf("red's child (%d -> %d) must not be red", key, right)
			}
		}

		return nil
	})
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestInsertRbTreeRandom(t *testing.T) {
	rand.Seed(time.Now().Unix())

	size := 1000000
	keys := make([]int, size+7)

	// Initialize a Red-Black Tree with some values
	tree := rbtree.NewIntRbTreeWithCache(size, true)
	for i := 0; i < size; i++ {
		keys[i] = rand.Intn(size)
		err := tree.Put(keys[i], rand.Intn(size))
		if err != nil {
			t.Errorf("put(%d) faces error: %v", keys[i], err)
		}

		_, err = tree.Get(keys[i])
		if err != nil {
			t.Errorf("insert(%d) not works", keys[i])
			break
		}
	}

	err := tree.Iterate(func(key interface{}) error {
		left, _ := tree.Left(key)
		right, _ := tree.Right(key)

		color, err := tree.Color(key)
		if err != nil {
			tree.Lookup(
				key,
				true,
			)
			t.Errorf("get color fails (%d): %v", key, err)
			return err
		}

		if color == rbtree.Red {
			color, err = tree.Color(left)
			if err != nil {
				t.Errorf("get color fails: %v", err)
				return err
			}
			if color == rbtree.Red {
				t.Errorf("red's child (%d -> %d) must not be red", key, left)
			}

			color, err = tree.Color(right)
			if err != nil {
				t.Errorf("get color fails: %v", err)
				return err
			}
			if color == rbtree.Red {
				t.Errorf("red's child (%d -> %d) must not be red", key, right)
			}
		}

		return nil
	})
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestDeleteRbTreeRandom(t *testing.T) {
	rand.Seed(time.Now().Unix())

	size := 100000
	keys := make([]int, size+7)
	stats := make([]bool, size+7)

	// Initialize a Red-Black Tree with some values
	tree := rbtree.NewIntRbTreeWithCache(size, true)
	for i := 0; i < size; i++ {
		keys[i] = rand.Intn(size)
		err := tree.Put(keys[i], rand.Intn(size))
		if err != nil {
			t.Errorf("put(%d) faces error: %v", keys[i], err)
		}
	}

	deleted := make(map[int]bool)

	for i := 0; i < size; i++ {
		j := rand.Intn(size)
		if stats[j] {
			continue
		}

		err := tree.Delete(keys[j])
		if err != nil {
			if _, ok := deleted[keys[j]]; !ok {
				tree.Lookup(keys[j], true)
				t.Errorf("delete(%d) fails: %v", keys[j], err)
			}
		}

		_, err = tree.Get(keys[j])
		if err == nil {
			t.Errorf("delete(%d) not works", keys[j])
		}
		stats[j] = true
		deleted[keys[j]] = true
	}

	err := tree.Iterate(func(key interface{}) error {
		left, _ := tree.Left(key)
		right, _ := tree.Right(key)

		color, err := tree.Color(key)
		if err != nil {
			tree.Lookup(
				key,
				true,
			)
			t.Errorf("get color fails (%d): %v", key, err)
			return err
		}

		if color == rbtree.Red {
			color, err = tree.Color(left)
			if err != nil {
				t.Errorf("get color fails: %v", err)
				return err
			}
			if color == rbtree.Red {
				t.Errorf("red's child (%d -> %d) must not be red", key, left)
			}

			color, err = tree.Color(right)
			if err != nil {
				t.Errorf("get color fails: %v", err)
				return err
			}
			if color == rbtree.Red {
				t.Errorf("red's child (%d -> %d) must not be red", key, right)
				tree.Lookup(right, true)
			}
		}

		return nil
	})
	if err != nil {
		t.Errorf("%v", err)
	}
}

var rbTreeTable = []struct {
	input int
}{
	{input: 256},
	{input: 1024},
	{input: 10240},
}

func BenchmarkInsertRbTree(b *testing.B) {
	rand.Seed(time.Now().Unix())

	size := 1024000
	keys := make([]int, size+7)
	for i := 0; i < size; i++ {
		keys[i] = rand.Intn(size)
	}

	for _, v := range rbTreeTable {
		b.Run(
			fmt.Sprintf("input_size_%d", v.input),
			func(b *testing.B) {
				size := v.input

				// Initialize a Red-Black Tree with some values
				tree := rbtree.NewIntRbTreeWithCache(size, true)
				for i := 0; i < size; i++ {
					err := tree.Put(keys[i], keys[i])
					if err != nil {
						b.Errorf("put(%d) faces error: %v", keys[i], err)
					}
				}
			},
		)
	}
}

func BenchmarkDeleteRbTree(b *testing.B) {
	rand.Seed(time.Now().Unix())

	size := 1024000
	keys := make([]int, size+7)
	for i := 0; i < size; i++ {
		keys[i] = rand.Intn(size)
	}

	for _, v := range rbTreeTable {
		b.Run(
			fmt.Sprintf("input_size_%d", v.input),
			func(b *testing.B) {
				size := v.input

				// Initialize a Red-Black Tree with some values
				tree := rbtree.NewIntRbTreeWithCache(size, true)
				for i := 0; i < size; i++ {
					err := tree.Put(keys[i], keys[i])
					if err != nil {
						b.Errorf("put(%d) faces error: %v", keys[i], err)
					}
				}

				// Delete every node of this red-black tree
				for i := 0; i < size; i++ {
					tree.Delete(keys[i])
				}
			},
		)
	}
}
