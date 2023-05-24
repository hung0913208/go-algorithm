package rbtree

import (
	"errors"
	"log"

	"github.com/hung0913208/go-algorithm/lib/algorithm/heap"
)

type RbTree interface {
	// @NOTE: interact with the red-black tree
	Put(key, val interface{}) error
	Get(key interface{}) (interface{}, error)
	Delete(key interface{}) error
	Iterate(handler func(key interface{}) error) error

	// @NOTE: get information about the red-black tree
	Root() (interface{}, error)
	Size() int

	// @NOTE: just for debuging issues
	Debug(flag bool)
	Dump()
	Lookup(key interface{}, debug ...bool) error

	// @NOTE: get information about specific node by key
	Color(key interface{}) (Color, error)
	Left(key interface{}) (interface{}, error)
	Right(key interface{}) (interface{}, error)
}

type rbTreeImpl struct {
	compare heap.Comparator
	array   []rbNodeImpl
	root    int
	size    int
	gc      heap.Heap
	debug   bool
}

type rbNodeImpl struct {
	key, val interface{}
	index    int
	parent   int
	childs   []int
	color    Color
	dead     bool
}

type Color int

const (
	Black Color = iota
	Red
)

const (
	EMPTY = 0
	SKIP  = 2
	LEFT  = 0
	RIGHT = 1
)

func NewRbTreeWithCacheSize(size int, compare Comparator) RbTree {
	return &rbTreeImpl{
		gc:      NewIntHeap(true),
		root:    EMPTY,
		size:    1,
		array:   make([]rbNodeImpl, size+1),
		compare: compare,
	}
}

func NewIntRbTree(isAsc bool) RbTree {
	return NewIntRbTreeWithCache(0, isAsc)
}

func NewIntRbTreeWithCache(size int, isAsc bool) RbTree {
	return NewRbTreeWithCacheSize(
		size,
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

func (self *rbTreeImpl) Debug(flag bool) {
	self.debug = flag
}

func (self *rbTreeImpl) Get(key interface{}) (interface{}, error) {
	var ret interface{}

	return ret, self.lookup(key, func(index, bnext int) error {
		ret = self.array[index].val
		return nil
	})
}

func (self *rbTreeImpl) Put(key, val interface{}) error {
	if self.root == EMPTY {
		index, _, err := self.allocate(key, val, Black)
		if err != nil {
			return err
		}

		self.root = index
		return nil
	}

	curr := self.root
	bnext := EMPTY

	for {
		compare := self.compare(self.array[curr].key, key)

		switch {
		case compare == 0:
			self.array[curr].val = val
			return nil

		case compare < 0:
			if self.array[curr].childs[LEFT] == EMPTY {
				child, _, err := self.allocate(key, val, Red)
				if err != nil {
					return err
				}

				self.array[curr].childs[LEFT] = child
				self.array[child].parent = curr
				return self.insertCase1(child)
			}

			bnext = LEFT

		default:
			if self.array[curr].childs[RIGHT] == EMPTY {
				child, _, err := self.allocate(key, val, Red)
				if err != nil {
					return err
				}

				self.array[curr].childs[RIGHT] = child
				self.array[child].parent = curr
				return self.insertCase1(child)
			}

			bnext = RIGHT
		}

		curr = self.array[curr].childs[bnext]
	}
}

func (self *rbTreeImpl) Delete(key interface{}) error {
	return self.lookup(key, func(node, turn int) error {
		left := self.array[node].childs[LEFT]
		right := self.array[node].childs[RIGHT]
		remove := node
		replace := node

		if left != EMPTY && right != EMPTY {
			remove = self.farthestRight(left)

			if self.debug {
				log.Print("bot left and right is not empty: ",
					" remove=", self.valueOf(remove),
					" replace=", self.valueOf(replace))
			}

			self.array[replace].key = self.array[remove].key
			self.array[replace].val = self.array[remove].val

			left = self.array[remove].childs[LEFT]
			right = self.array[remove].childs[RIGHT]
		}

		child := left
		if child == EMPTY {
			child = right
		}

		if self.array[remove].color == Black {
			self.array[remove].color = self.array[child].color
			self.deleteCase1(remove)
		}

		// @NOTE: replace_node(remove, child)
		parent := self.array[remove].parent

		if parent == EMPTY {
			self.root = child
		} else if remove == self.array[parent].childs[LEFT] {
			self.array[parent].childs[LEFT] = child
		} else {
			self.array[parent].childs[RIGHT] = child
		}

		if child != EMPTY {
			self.array[child].parent = parent
		}

		// @NOTE: detach unused nodes
		if parent == EMPTY && child != EMPTY {
			self.array[child].color = Black
		}

		if err := self.gc.Push(remove); err != nil {
			return err
		}

		self.array[remove].dead = true
		return nil
	})
}

func (self *rbTreeImpl) Iterate(handler func(key interface{}) error) error {
	if len(self.array) == 0 {
		return errors.New("can't iterate an empty tree")
	}

	for _, item := range self.array {
		if item.index == EMPTY || item.dead {
			continue
		}

		if err := handler(item.key); err != nil {
			return err
		}
	}

	return nil
}

func (self *rbTreeImpl) Root() (interface{}, error) {
	if self.root == EMPTY {
		return nil, errors.New("The tree is empty")
	}

	return self.array[self.root].key, nil
}

func (self *rbTreeImpl) Size() int {
	return self.size - (EMPTY + 1)
}

func (self *rbTreeImpl) Dump() {
	log.Printf("root=%d", self.root)
	for _, item := range self.array {
		log.Printf(
			"index=%d key=%d parent=%d left=%d right=%d color=%d dead=%v",
			item.index,
			item.key,
			item.parent,
			item.childs[LEFT],
			item.childs[RIGHT],
			item.color,
			item.dead,
		)
	}
}

func (self *rbTreeImpl) Lookup(
	key interface{},
	debug ...bool,
) error {
	return self.lookup(
		key,
		func(index, direct int) error {
			return nil
		},
		debug...,
	)
}

func (self *rbTreeImpl) Color(key interface{}) (Color, error) {
	var color Color

	return color, self.lookup(key, func(index, turn int) error {
		if index == EMPTY {
			color = Black
			return nil
		}

		color = self.array[index].color
		return nil
	})
}

func (self *rbTreeImpl) Left(key interface{}) (interface{}, error) {
	var node *rbNodeImpl

	err := self.lookup(key, func(index, turn int) error {
		left := self.array[index].childs[LEFT]
		node = &self.array[left]

		if left == EMPTY {
			return errors.New("not found left branch")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return node.key, nil
}

func (self *rbTreeImpl) Right(key interface{}) (interface{}, error) {
	var node *rbNodeImpl

	err := self.lookup(key, func(index, turn int) error {
		right := self.array[index].childs[RIGHT]
		node = &self.array[right]

		if right == EMPTY {
			return errors.New("not found right branch")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return node.key, nil
}

func (self *rbTreeImpl) allocate(
	key, val interface{},
	color Color,
) (int, *rbNodeImpl, error) {
	pickingIndex := self.size

	if len(self.array[0].childs) == 0 {
		self.array[0].childs = []int{EMPTY, EMPTY}
	}

	if len(self.array) == self.size {
		pickingIndex = self.size

		self.array = append(
			self.array,
			rbNodeImpl{
				key:    key,
				val:    val,
				index:  self.size,
				color:  color,
				parent: EMPTY,
				childs: []int{EMPTY, EMPTY},
			},
		)
		self.size++
	} else if self.gc.Size() > 0 {
		index, err := self.gc.Get()
		if err != nil {
			return 0, nil, err
		}

		err = self.gc.Pop()
		if err != nil {
			return 0, nil, err
		}

		pickingIndex = index.(int)
	} else {
		self.array[self.size] = rbNodeImpl{
			key:    key,
			val:    val,
			index:  self.size,
			color:  color,
			parent: EMPTY,
			childs: []int{EMPTY, EMPTY},
		}
		self.size++
	}

	self.array[pickingIndex].color = color
	self.array[pickingIndex].dead = false
	self.array[pickingIndex].key = key
	self.array[pickingIndex].val = val

	return pickingIndex, &self.array[pickingIndex], nil
}

func (self *rbTreeImpl) lookup(
	key interface{},
	handler func(index, direct int) error,
	debug ...bool,
) error {
	if key == nil {
		return handler(EMPTY, LEFT)
	}

	if self.root == EMPTY {
		return errors.New("not found the key")
	}

	curr := self.root
	bnext := EMPTY

	for {
		compare := self.compare(self.array[curr].key, key)

		if len(debug) > 0 && debug[0] {
			log.Printf(
				"index=%d key=%d left=%d right=%d color=%d dead=%v compare=%d",
				self.array[curr].index,
				self.array[curr].key,
				self.array[curr].childs[LEFT],
				self.array[curr].childs[RIGHT],
				self.array[curr].color,
				self.array[curr].dead,
				compare,
			)
		}

		switch {
		case compare == 0:
			if self.array[curr].dead {
				return errors.New("key has been deleted")
			}
			return handler(curr, bnext)

		case compare < 0:
			bnext = LEFT

		default:
			bnext = RIGHT
		}

		if self.array[curr].childs[bnext] == EMPTY {
			return errors.New("not found the key")
		}

		curr = self.array[curr].childs[bnext]
	}
}

func (self *rbTreeImpl) insertCase1(me int) error {
	if self.array[me].parent == EMPTY {
		if self.debug {
			log.Print("root to black ", self.valueOf(me))
		}

		self.array[me].color = Black
		return nil
	}
	return self.insertCase2(me)
}

func (self *rbTreeImpl) insertCase2(me int) error {
	parent := self.array[me].parent
	rotate := LEFT

	if self.array[parent].color == Black {
		return nil
	}

	if self.array[parent].childs[RIGHT] == me {
		rotate = RIGHT
	}

	return self.insertCase3(me, parent, rotate)
}

func (self *rbTreeImpl) insertCase3(me, parent, rotate int) error {
	var err error

	grant := self.array[parent].parent
	uncle := self.array[grant].childs[LEFT]
	swing := false

	if uncle == parent {
		// left - left
		uncle = self.array[grant].childs[RIGHT]

		if rotate != LEFT {
			// left - right
			rotate = LEFT
			swing = true
		}
	} else if rotate != RIGHT {
		// right - left
		rotate = RIGHT
		swing = true
	} // else: right - right

	if self.array[uncle].color == Red {
		// @NOTE: recolor parent and uncle to black
		if self.debug {
			log.Print("- recolor ",
				self.valueOf(me),
				self.valueOf(parent),
				self.valueOf(uncle),
				self.valueOf(grant))
			self.Dump()
		}
		self.array[parent].color = Black
		self.array[uncle].color = Black
		self.array[grant].color = Red

		if self.debug {
			log.Print("+ recolor ",
				self.valueOf(me),
				self.valueOf(parent),
				self.valueOf(uncle),
				self.valueOf(grant))
			self.Dump()
		}

		return self.insertCase1(grant)
	}

	if swing {
		if self.debug {
			log.Print("- case 4: ",
				self.valueOf(me),
				self.valueOf(parent),
				self.valueOf(uncle),
				self.valueOf(grant),
				" rotate=", rotate == LEFT)
			self.Dump()
		}

		me, err = self.insertCase4(me, parent, rotate)
		if err != nil {
			return err
		}

		parent = self.array[me].parent
		grant = self.array[parent].parent
		uncle = self.array[grant].childs[LEFT]
		if uncle == parent {
			uncle = self.array[grant].childs[RIGHT]
		}

		if self.debug {
			log.Print("+ case 4: ",
				self.valueOf(me),
				self.valueOf(parent),
				self.valueOf(uncle),
				self.valueOf(grant),
				" rotate=", rotate == LEFT)
			self.Dump()
		}

	}

	if rotate == LEFT {
		rotate = RIGHT
	} else {
		rotate = LEFT
	}

	if self.debug {
		log.Print("- case 5: ",
			self.valueOf(me),
			self.valueOf(parent),
			self.valueOf(uncle),
			self.valueOf(grant),
			" swing=", swing,
			" rotate=", rotate == LEFT)
		self.Dump()
	}

	self.insertCase5(me, parent, grant, rotate)
	if self.debug {
		log.Print("+ case 5: ",
			self.valueOf(me),
			self.valueOf(parent),
			self.valueOf(uncle),
			self.valueOf(grant),
			" swing=", swing,
			" rotate=", rotate == LEFT)
		self.Dump()
	}
	return nil
}

func (self *rbTreeImpl) insertCase4(
	me, parent, rotate int,
) (int, error) {
	var err error

	switch rotate {
	case LEFT:
		parent, err = self.rotateLeft(parent)
		if err != nil {
			return EMPTY, err
		}

		return self.array[me].childs[LEFT], nil

	case RIGHT:
		parent, err = self.rotateRight(parent)
		if err != nil {
			return EMPTY, err
		}

		return self.array[me].childs[RIGHT], nil

	default:
		return EMPTY, errors.New("unknown rotation")
	}
}

func (self *rbTreeImpl) insertCase5(
	me, parent, grant, rotate int,
) error {
	var err error

	self.array[parent].color = Black
	if grant != EMPTY {
		self.array[grant].color = Red

		switch rotate {
		case LEFT:
			grant, err = self.rotateLeft(grant)
			if err != nil {
				return err
			}

		case RIGHT:
			grant, err = self.rotateRight(grant)
			if err != nil {
				return err
			}

		default:
			return errors.New("unknown rotation")
		}
	}

	return nil
}

func (self *rbTreeImpl) deleteCase1(me int) error {
	if self.debug {
		log.Print("delete case 1:", " me=", self.valueOf(me))
	}

	parent := self.array[me].parent
	if parent == EMPTY {
		return nil
	}

	return self.deleteCase2(me, parent)
}

func (self *rbTreeImpl) deleteCase2(me, parent int) error {
	sibling := self.array[parent].childs[LEFT]
	if sibling == me {
		sibling = self.array[parent].childs[RIGHT]
	}

	if self.debug {
		log.Print("delete case 2:",
			" me=", self.valueOf(me),
			" parent=", self.valueOf(parent),
			" sibling=", self.valueOf(sibling))
	}

	if sibling == EMPTY {
		return self.deleteCase1(parent)
	}

	if self.array[sibling].color == Red {
		var err error

		if self.debug {
			log.Print("red sibling, exchange color of sibling and parent and rotate parent")
		}

		self.array[parent].color = Red
		self.array[sibling].color = Black

		if me == self.array[parent].childs[LEFT] {
			_, err = self.rotateLeft(parent)
		} else {
			_, err = self.rotateRight(parent)
		}
		if err != nil {
			return err
		}

		parent = self.array[me].parent
		sibling = self.array[parent].childs[LEFT]

		if sibling == me {
			sibling = self.array[parent].childs[RIGHT]
		}
	}

	return self.deleteCase3(me, parent, sibling)
}

func (self *rbTreeImpl) deleteCase3(me, parent, sibling int) error {
	left := self.array[sibling].childs[LEFT]
	right := self.array[sibling].childs[RIGHT]

	if self.debug {
		log.Print("delete case 3:",
			" me=", self.valueOf(me),
			" parent=", self.valueOf(parent),
			" sibling=", self.valueOf(sibling),
			" left=", self.valueOf(left),
			" right=", self.valueOf(right),
		)
	}

	if self.array[parent].color == Black &&
		self.array[sibling].color == Black &&
		self.array[left].color == Black &&
		self.array[right].color == Black {
		self.array[sibling].color = Red

		if self.debug {
			log.Print("all node are black, switch sibling to red")
		}
		return self.deleteCase1(parent)
	}

	return self.deleteCase4(me, parent, sibling, left, right)
}

func (self *rbTreeImpl) deleteCase4(
	me, parent, sibling, left, right int,
) error {
	if self.debug {
		log.Print("delete case 4:",
			" me=", self.valueOf(me),
			" parent=", self.valueOf(parent),
			" sibling=", self.valueOf(sibling),
			" left=", self.valueOf(left),
			" right=", self.valueOf(right),
		)
	}

	if self.array[parent].color == Red &&
		self.array[sibling].color == Black &&
		self.array[left].color == Black &&
		self.array[right].color == Black {
		self.array[sibling].color = Red
		self.array[parent].color = Black

		if self.debug {
			log.Print("only parent is red, swap color between sibling and parent")
		}
	} else {
		return self.deleteCase5(me, parent, sibling, left, right)
	}

	return nil
}

func (self *rbTreeImpl) deleteCase5(
	me, parent, sibling, left, right int,
) error {
	var err error

	if self.debug {
		log.Print("delete case 5:",
			" me=", self.valueOf(me),
			" parent=", self.valueOf(parent),
			" sibling=", self.valueOf(sibling),
			" left=", self.valueOf(left),
			" right=", self.valueOf(right),
		)
	}

	if me == self.array[parent].childs[LEFT] &&
		self.array[sibling].color == Black &&
		self.array[left].color == Red &&
		self.array[right].color == Black {

		if self.debug {
			log.Print("rotate sibling and left node")
		}

		self.array[sibling].color = Red
		self.array[left].color = Black
		sibling, err = self.rotateRight(sibling)
		if err != nil {
			return err
		}
	} else if me == self.array[parent].childs[RIGHT] &&
		self.array[sibling].color == Black &&
		self.array[right].color == Red &&
		self.array[left].color == Black {

		if self.debug {
			log.Print("rotate sibling and right node")
		}

		self.array[sibling].color = Red
		self.array[right].color = Black
		sibling, err = self.rotateLeft(sibling)
		if err != nil {
			return err
		}

		if self.debug {
			self.Dump()
		}
	}

	left = self.array[sibling].childs[LEFT]
	right = self.array[sibling].childs[RIGHT]

	return self.deleteCase6(me, parent, sibling, left, right)
}

func (self *rbTreeImpl) deleteCase6(
	me, parent, sibling, left, right int,
) error {
	var err error

	if self.debug {
		log.Print("delete case 6:",
			" me=", self.valueOf(me),
			" parent=", self.valueOf(parent),
			" sibling=", self.valueOf(sibling),
			" left=", self.valueOf(left),
			" right=", self.valueOf(right),
		)
	}

	self.array[sibling].color = self.array[parent].color
	self.array[parent].color = Black

	if me == self.array[parent].childs[LEFT] &&
		self.array[right].color == Red {
		self.array[right].color = Black
		if self.debug {
			log.Print("rotate left parent and change left node")
		}

		_, err = self.rotateLeft(parent)
	} else if self.array[left].color == Red {
		if self.debug {
			log.Print("rotate right parent and change left node")
		}

		self.array[left].color = Black
		_, err = self.rotateRight(parent)
	}

	return err
}

func (self *rbTreeImpl) farthestRight(me int) int {
	for me != EMPTY {
		if self.array[me].childs[RIGHT] == EMPTY {
			return me
		}

		me = self.array[me].childs[RIGHT]
	}
	return EMPTY
}

func (self *rbTreeImpl) farthestLeft(me int) int {
	for me != EMPTY {
		if self.array[me].childs[LEFT] == EMPTY {
			return me
		}

		me = self.array[me].childs[LEFT]
	}
	return EMPTY
}

func (self *rbTreeImpl) rotateLeft(root int) (int, error) {
	parent := self.array[root].parent
	right := self.array[root].childs[RIGHT]
	rightLeft := self.array[right].childs[LEFT]

	if self.debug {
		log.Print("rotate left:",
			" root=", self.valueOf(root),
			" parent=", self.valueOf(parent),
			" right=", self.valueOf(right),
			" right->left=", self.valueOf(rightLeft))
	}

	if parent == EMPTY {
		self.root = right
	} else if root == self.array[parent].childs[LEFT] {
		self.array[parent].childs[LEFT] = right
	} else {
		self.array[parent].childs[RIGHT] = right
	}

	if right != EMPTY {
		self.array[right].parent = parent
		self.array[right].childs[LEFT] = root
	}

	if rightLeft != EMPTY {
		self.array[rightLeft].parent = root
	}

	self.array[root].parent = right
	self.array[root].childs[RIGHT] = rightLeft
	return right, nil
}

func (self *rbTreeImpl) rotateRight(root int) (int, error) {
	parent := self.array[root].parent
	left := self.array[root].childs[LEFT]
	leftRight := self.array[left].childs[RIGHT]

	if self.debug {
		log.Print("rotate right: ",
			" root=", self.valueOf(root),
			" parent=", self.valueOf(parent),
			" left=", self.valueOf(left),
			" left->right=", self.valueOf(leftRight))
	}

	if parent == EMPTY {
		self.root = left
	} else if root == self.array[parent].childs[LEFT] {
		self.array[parent].childs[LEFT] = left
	} else {
		self.array[parent].childs[RIGHT] = left
	}

	if left != EMPTY {
		self.array[left].parent = parent
		self.array[left].childs[RIGHT] = root
	}

	if leftRight != EMPTY {
		self.array[leftRight].parent = root
	}

	self.array[root].parent = left
	self.array[root].childs[LEFT] = leftRight
	return left, nil
}

func (self *rbTreeImpl) valueOf(index int) interface{} {
	return self.array[index].key
}
