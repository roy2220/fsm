package list_test

import (
	"fmt"
	"testing"

	"github.com/roy2220/fsm/internal/list"
	"github.com/stretchr/testify/assert"
)

func TestListInsertItem(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize)
	l := new(list.List).Init()
	l.PrependItem(sa, 3*list.ItemSize)
	l.PrependItem(sa, 2*list.ItemSize)
	l.PrependItem(sa, 1*list.ItemSize)
	l.AppendItem(sa, 4*list.ItemSize)
	l.AppendItem(sa, 5*list.ItemSize)
	l.AppendItem(sa, 6*list.ItemSize)
	assert.Equal(t, "1,2,3,4,5,6", DumpList(sa, l))
}

func TestListRemoveItem(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize)
	l := new(list.List).Init()

	for n := 1; n <= 6; n++ {
		l.AppendItem(sa, int64(n*list.ItemSize))
	}

	assert.Equal(t, "1,2,3,4,5,6", DumpList(sa, l))
	getItem := l.GetItems()

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		if n := i / list.ItemSize; n%2 == 0 {
			l.RemoveItem(sa, i)
			l.AppendItem(sa, i)
		}
	}

	assert.Equal(t, "1,3,5,2,4,6", DumpList(sa, l))
	getItem = l.GetItems()

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		if n := i / list.ItemSize; n%2 == 0 {
			l.RemoveItem(sa, i)
			l.PrependItem(sa, i)
		}
	}

	assert.Equal(t, "6,4,2,1,3,5", DumpList(sa, l))
	getItem = l.GetItems()

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		l.RemoveItem(sa, i)
	}

	assert.Equal(t, "", DumpList(sa, l))
}

func TestListSetHeadAndTail(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize)
	l := new(list.List).Init()

	for n := 1; n <= 6; n++ {
		l.AppendItem(sa, int64(n*list.ItemSize))
	}

	assert.Equal(t, "1,2,3,4,5,6", DumpList(sa, l))
	l.SetHead(sa, 3*list.ItemSize)
	assert.Equal(t, "3,4,5,6,1,2", DumpList(sa, l))
	l.SetTail(sa, 5*list.ItemSize)
	assert.Equal(t, "6,1,2,3,4,5", DumpList(sa, l))

	for n := 1; n <= 5; n++ {
		l.RemoveItem(sa, int64(n*list.ItemSize))
	}

	assert.Equal(t, "6", DumpList(sa, l))
	l.SetHead(sa, 6*list.ItemSize)
	assert.Equal(t, "6", DumpList(sa, l))
	l.SetTail(sa, 6*list.ItemSize)
	assert.Equal(t, "6", DumpList(sa, l))
}

func TestListStoreAndLoad(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize)
	l := new(list.List).Init()

	for n := 1; n <= 6; n++ {
		l.AppendItem(sa, int64(n*list.ItemSize))
	}

	assert.Equal(t, "1,2,3,4,5,6", DumpList(sa, l))
	b := [list.Size]byte{}
	l.Store(&b)
	l.Load(&b)
	assert.Equal(t, "1,2,3,4,5,6", DumpList(sa, l))
}

func DumpList(sa []byte, l *list.List) string {
	getItem := l.GetItems()
	text := ""

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		text += fmt.Sprintf("%d,", int(i/list.ItemSize))
	}

	if text != "" {
		text = text[:len(text)-1]
	}

	return text
}
