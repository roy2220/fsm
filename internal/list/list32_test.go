package list_test

import (
	"fmt"
	"testing"

	"github.com/roy2220/fsm/internal/list"
	"github.com/stretchr/testify/assert"
)

func TestList32InsertItem(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize32)
	l := new(list.List32).Init()
	l.PrependItem(sa, 3*list.ItemSize32)
	l.PrependItem(sa, 2*list.ItemSize32)
	l.InsertItemBefore(sa, 1*list.ItemSize32, 2*list.ItemSize32)
	l.AppendItem(sa, 4*list.ItemSize32)
	l.AppendItem(sa, 5*list.ItemSize32)
	l.InsertItemAfter(sa, 6*list.ItemSize32, 5*list.ItemSize32)
	assert.Equal(t, "1,2,3,4,5,6", DumpList32(sa, l))
}

func TestList32RemoveItem(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize32)
	l := new(list.List32).Init()

	for n := 1; n <= 6; n++ {
		l.AppendItem(sa, int32(n*list.ItemSize32))
	}

	assert.Equal(t, "1,2,3,4,5,6", DumpList32(sa, l))
	getItem := l.GetItems()

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		if n := i / list.ItemSize32; n%2 == 0 {
			l.RemoveItem(sa, i)
			l.AppendItem(sa, i)
		}
	}

	assert.Equal(t, "1,3,5,2,4,6", DumpList32(sa, l))
	getItem = l.GetItems()

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		if n := i / list.ItemSize32; n%2 == 0 {
			l.RemoveItem(sa, i)
			l.PrependItem(sa, i)
		}
	}

	assert.Equal(t, "6,4,2,1,3,5", DumpList32(sa, l))
	getItem = l.GetItems()

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		l.RemoveItem(sa, i)
	}

	assert.Equal(t, "", DumpList32(sa, l))
}

func TestList32SetHeadAndTail(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize32)
	l := new(list.List32).Init()

	for n := 1; n <= 6; n++ {
		l.AppendItem(sa, int32(n*list.ItemSize32))
	}

	assert.Equal(t, "1,2,3,4,5,6", DumpList32(sa, l))
	l.SetHead(sa, 3*list.ItemSize32)
	assert.Equal(t, "3,4,5,6,1,2", DumpList32(sa, l))
	l.SetTail(sa, 5*list.ItemSize32)
	assert.Equal(t, "6,1,2,3,4,5", DumpList32(sa, l))

	for n := 1; n <= 5; n++ {
		l.RemoveItem(sa, int32(n*list.ItemSize32))
	}

	assert.Equal(t, "6", DumpList32(sa, l))
	l.SetHead(sa, 6*list.ItemSize32)
	assert.Equal(t, "6", DumpList32(sa, l))
	l.SetTail(sa, 6*list.ItemSize32)
	assert.Equal(t, "6", DumpList32(sa, l))
}

func TestList32StoreAndLoad(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize32)
	l := new(list.List32).Init()

	for n := 1; n <= 6; n++ {
		l.AppendItem(sa, int32(n*list.ItemSize32))
	}

	assert.Equal(t, "1,2,3,4,5,6", DumpList32(sa, l))
	b := [list.Size32]byte{}
	l.Store(b[:])
	l.Load(b[:])
	assert.Equal(t, "1,2,3,4,5,6", DumpList32(sa, l))
}

func TestList32Flags(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize32)
	l := new(list.List32).Init()

	for n := 1; n <= 6; n++ {
		l.AppendItem(sa, int32(n*list.ItemSize32))
		list.SetItem32Flags(sa, int32(n*list.ItemSize32), int8(n%4))
	}

	assert.Equal(t, "1,2,3,4,5,6", DumpList32(sa, l))
	getItem := l.GetItems()

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		assert.Equal(t, int8(i/list.ItemSize32%4), list.Item32Flags(sa, i))
	}
}

func DumpList32(sa []byte, l *list.List32) string {
	getItem := l.GetItems()
	text := ""

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		text += fmt.Sprintf("%d,", int(i/list.ItemSize32))
	}

	if text != "" {
		text = text[:len(text)-1]
	}

	return text
}
