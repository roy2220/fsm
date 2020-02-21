package rbtree_test

import (
	"math/rand"
	"testing"

	"github.com/roy2220/fsm/internal/rbtree"
	"github.com/stretchr/testify/assert"
)

func TestRBTreeDeleteKey(t *testing.T) {
	rbt, ks := MakeRBTree()

	for _, k := range ks {
		ok := rbt.DeleteKey(k)
		assert.True(t, ok, "%v", k)
	}

	ok := rbt.DeleteKey(0)
	assert.False(t, ok)
}

func TestRBTreeDeleteMinKey(t *testing.T) {
	rbt, ks := MakeRBTree()

	for i, k := range ks {
		mk, ok := rbt.DeleteMinKey()

		if assert.True(t, ok, "%v", k) {
			assert.Equal(t, int64(i), mk)
		}
	}

	_, ok := rbt.DeleteMinKey()
	assert.False(t, ok)
}

func TestRBTreeDeleteMaxKey(t *testing.T) {
	rbt, ks := MakeRBTree()

	for i, k := range ks {
		mk, ok := rbt.DeleteMaxKey()

		if assert.True(t, ok, "%v", k) {
			assert.Equal(t, int64(len(ks)-1-i), mk)
		}
	}

	_, ok := rbt.DeleteMaxKey()
	assert.False(t, ok)
}

func TestRBTreeFindKey(t *testing.T) {
	rbt, ks := MakeRBTree()

	for _, k := range ks {
		ok := rbt.FindKey(k)
		assert.True(t, ok, "%v", k)
	}
}

func TestRBTreeGetKeys(t *testing.T) {
	rbt, ks := MakeRBTree()
	getKey := rbt.GetKeys()
	i := 0

	for k, ok := getKey(); ok; k, ok = getKey() {
		assert.Equal(t, int64(i), k, "%v", k)
		i++
	}

	assert.Equal(t, len(ks), i)
}

func MakeRBTree() (*rbtree.RBTree, []int64) {
	ks := make([]int64, 100000)

	for i := range ks {
		ks[i] = int64(i)
	}

	rand.Shuffle(len(ks), func(i, j int) {
		ks[i], ks[j] = ks[j], ks[i]
	})

	rbt := new(rbtree.RBTree).Init()

	for _, k := range ks {
		rbt.AddKey(k)
	}

	return rbt, ks
}
