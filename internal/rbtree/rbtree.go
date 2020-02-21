// Package rbtree implements a red-black tree.
package rbtree

import (
	"errors"
	"fmt"
	"io"
)

// RBTree represents a red-black tree.
type RBTree struct {
	nil node
}

// Init initializes the red-black tree and returns it.
func (rbt *RBTree) Init() *RBTree {
	rbt.nil.SetColor(black)
	rbt.setRoot(&rbt.nil)
	return rbt
}

// AddKey inserts the given key to the red-black tree.
func (rbt *RBTree) AddKey(key int64) {
	rbt.insertNode(&node{
		LeftChild:   &rbt.nil,
		RightChild:  &rbt.nil,
		ColorAndKey: makeColorAndKey(red, key),
	})
}

// DeleteKey deletes the given key from the red-black tree and returns true
// if the given key exists, otherwise it returns false.
func (rbt *RBTree) DeleteKey(key int64) bool {
	x := rbt.root()

	for x != &rbt.nil {
		d := key - x.Key()

		if d == 0 {
			break
		}

		if d < 0 {
			x = x.LeftChild
		} else {
			x = x.RightChild
		}
	}

	if x == &rbt.nil {
		return false
	}

	rbt.removeNode(x)
	return true
}

// DeleteMinKey deletes the minimum key from the red-black tree and returns true
// if the tree is not empty, otherwise it returns false.
func (rbt *RBTree) DeleteMinKey() (int64, bool) {
	x := rbt.root()

	if x == &rbt.nil {
		return 0, false
	}

	for ; x.LeftChild != &rbt.nil; x = x.LeftChild {
	}

	key := x.Key()
	rbt.removeNode(x)
	return key, true
}

// DeleteMaxKey deletes the maximum key from the red-black tree and returns true
// if the tree is not empty, otherwise it returns false.
func (rbt *RBTree) DeleteMaxKey() (int64, bool) {
	x := rbt.root()

	if x == &rbt.nil {
		return 0, false
	}

	for ; x.RightChild != &rbt.nil; x = x.RightChild {
	}

	key := x.Key()
	rbt.removeNode(x)
	return key, true
}

// FindKey finds the given key in the red-black tree and returns true
// if the given key exists otherwise it returns false.
func (rbt *RBTree) FindKey(key int64) bool {
	x := rbt.root()

	for x != &rbt.nil {
		d := key - x.Key()

		if d == 0 {
			return true
		}

		if d < 0 {
			x = x.LeftChild
		} else {
			x = x.RightChild
		}
	}

	return false
}

// GetKeys returns an iteration function to iterate over all keys in the red-black tree.
func (rbt *RBTree) GetKeys() func() (int64, bool) {
	stack := []*node(nil)

	for x := rbt.root(); x != &rbt.nil; x = x.LeftChild {
		stack = append(stack, x)

	}

	return func() (int64, bool) {
		i := len(stack) - 1

		if i < 0 {
			return 0, false
		}

		x := stack[i]
		stack = stack[:i]

		for y := x.RightChild; y != &rbt.nil; y = y.LeftChild {
			stack = append(stack, y)
		}

		return x.Key(), true
	}
}

// Fprint dumps the red-black tree as plain text for debugging purposes
func (rbt *RBTree) Fprint(writer io.Writer) error {
	return rbt.doFprint(writer, rbt.root(), "\n")
}

func (rbt *RBTree) setRoot(root *node) {
	rbt.nil.SetLeftChild(root)
}

func (rbt *RBTree) insertNode(x *node) {
	y := &rbt.nil
	z, f := rbt.nil.LeftChild /* rbt.root() */, (*node).SetLeftChild /* rbt.setRoot */

	for z != &rbt.nil {
		y = z

		if x.Key() < y.Key() {
			z, f = y.LeftChild, (*node).SetLeftChild
		} else {
			z, f = y.RightChild, (*node).SetRightChild
		}
	}

	f(y, x)
	rbt.fixAfterNodeInsertion(x)
}

func (rbt *RBTree) fixAfterNodeInsertion(x *node) {
	for {
		y := x.Parent

		if y.Color() == black {
			break
		}

		z := y.Parent
		var v *node

		if y == z.LeftChild {
			v = z.RightChild

			if v.Color() == black {
				if x == y.RightChild {
					y.RotateLeft()
					x, y = y, x
				}

				y.SetColor(black)
				z.SetColor(red)
				z.RotateRight()
				break
			}
		} else {
			v = z.LeftChild

			if v.Color() == black {
				if x == y.LeftChild {
					y.RotateRight()
					x, y = y, x
				}

				y.SetColor(black)
				z.SetColor(red)
				z.RotateLeft()
				break
			}
		}

		y.SetColor(black)
		z.SetColor(red)
		v.SetColor(black)
		x = z
	}

	rbt.root().SetColor(black)
}

func (rbt *RBTree) removeNode(x *node) {
	var y, z *node

	if x.LeftChild == &rbt.nil {
		y, z = x, x.RightChild
	} else if x.RightChild == &rbt.nil {
		y, z = x, x.LeftChild
	} else {
		for v, w := x.LeftChild, x.RightChild; ; v, w = v.RightChild, w.LeftChild {
			if v.RightChild == &rbt.nil {
				y, z = v, v.LeftChild
				break
			}

			if w.LeftChild == &rbt.nil {
				y, z = w, w.RightChild
				break
			}
		}
	}

	y.Replace(z)

	if x != y {
		x.SetKey(y.Key())
	}

	if y.Color() == black {
		rbt.fixAfterNodeRemoval(z)
	}
}

func (rbt *RBTree) fixAfterNodeRemoval(x *node) {
	for x.Color() == black && x != rbt.root() {
		y := x.Parent
		var z *node

		if x == y.LeftChild {
			z = y.RightChild

			if z.Color() == red {
				y.SetColor(red)
				z.SetColor(black)
				y.RotateLeft()
				z = y.RightChild
			}

			v := z.RightChild
			w := z.LeftChild

			if v.Color() == red || w.Color() == red {
				if v.Color() == black {
					z.SetColor(red)
					w.SetColor(black)
					z.RotateRight()
					v, z = z, w
				}

				z.SetColor(y.Color())
				y.SetColor(black)
				v.SetColor(black)
				y.RotateLeft()
				x = rbt.root()
				break
			}
		} else {
			z = y.LeftChild

			if z.Color() == red {
				y.SetColor(red)
				z.SetColor(black)
				y.RotateRight()
				z = y.LeftChild
			}

			v := z.LeftChild
			w := z.RightChild

			if v.Color() == red || w.Color() == red {
				if v.Color() == black {
					z.SetColor(red)
					w.SetColor(black)
					z.RotateLeft()
					v, z = z, w
				}

				z.SetColor(y.Color())
				y.SetColor(black)
				v.SetColor(black)
				y.RotateRight()
				x = rbt.root()
				break
			}
		}

		z.SetColor(red)
		x = y
	}

	x.SetColor(black)
}

func (rbt *RBTree) doFprint(writer io.Writer, x *node, newLine string) error {
	if x == &rbt.nil {
		_, err := fmt.Fprint(writer, "<nil>")
		return err
	}

	var s string

	if x.Color() == red {
		s = "◯"
	} else {
		s = "●"
	}

	if _, err := fmt.Fprintf(writer, "%s %d", s, x.Key()); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(writer, "%s├──", newLine); err != nil {
		return err
	}

	if err := rbt.doFprint(writer, x.LeftChild, newLine+"│  "); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(writer, "%s└──", newLine); err != nil {
		return err
	}

	if err := rbt.doFprint(writer, x.RightChild, newLine+"   "); err != nil {
		return err
	}

	return nil
}

func (rbt *RBTree) root() *node {
	return rbt.nil.LeftChild
}

const (
	red   = 0
	black = 1
)

type node struct {
	Parent, LeftChild, RightChild *node
	ColorAndKey                   uint64
}

func (n *node) SetLeftChild(leftChild *node) {
	n.LeftChild = leftChild
	leftChild.Parent = n
}

func (n *node) SetRightChild(rightChild *node) {
	n.RightChild = rightChild
	rightChild.Parent = n
}

func (n *node) Replace(other *node) {
	if n == n.Parent.LeftChild {
		n.Parent.SetLeftChild(other)
	} else {
		n.Parent.SetRightChild(other)
	}
}

func (n *node) RotateLeft() {
	substitute := n.RightChild
	n.SetRightChild(substitute.LeftChild)
	n.Replace(substitute)
	substitute.SetLeftChild(n)
}

func (n *node) RotateRight() {
	substitute := n.LeftChild
	n.SetLeftChild(substitute.RightChild)
	n.Replace(substitute)
	substitute.SetRightChild(n)
}

func (n *node) SetColor(color color) {
	n.ColorAndKey = (uint64(color) << 63) | (n.ColorAndKey & 0x7FFFFFFFFFFFFFFF)
}

func (n *node) SetKey(key int64) {
	n.ColorAndKey = (n.ColorAndKey & 0x8000000000000000) | uint64(key)
}

func (n *node) Color() color {
	return color((n.ColorAndKey & 0x8000000000000000) >> 63)
}

func (n *node) Key() int64 {
	return int64(n.ColorAndKey & 0x7FFFFFFFFFFFFFFF)
}

type color int

func makeColorAndKey(color color, key int64) uint64 {
	if key < 0 {
		panic(errors.New("rbtree: invalid key"))
	}

	return (uint64(color) << 63) | uint64(key)
}
