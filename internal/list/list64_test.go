package list_test

import (
	"fmt"
	"testing"

	"github.com/roy2220/fsm/internal/list"
	"github.com/stretchr/testify/assert"
)

func TestList64InsertItem(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize64)
	l := new(list.List64).Init()
	l.PrependItem(sa, 3*list.ItemSize64)
	l.PrependItem(sa, 2*list.ItemSize64)
	l.InsertItemBefore(sa, 1*list.ItemSize64, 2*list.ItemSize64)
	l.AppendItem(sa, 4*list.ItemSize64)
	l.AppendItem(sa, 5*list.ItemSize64)
	l.InsertItemAfter(sa, 6*list.ItemSize64, 5*list.ItemSize64)
	assert.Equal(t, "1,2,3,4,5,6", DumpList64(sa, l))
}

func TestList64RemoveItem(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize64)
	l := new(list.List64).Init()

	for n := 1; n <= 6; n++ {
		l.AppendItem(sa, int64(n*list.ItemSize64))
	}

	assert.Equal(t, "1,2,3,4,5,6", DumpList64(sa, l))
	getItem := l.GetItems()

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		if n := i / list.ItemSize64; n%2 == 0 {
			l.RemoveItem(sa, i)
			l.AppendItem(sa, i)
		}
	}

	assert.Equal(t, "1,3,5,2,4,6", DumpList64(sa, l))
	getItem = l.GetItems()

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		if n := i / list.ItemSize64; n%2 == 0 {
			l.RemoveItem(sa, i)
			l.PrependItem(sa, i)
		}
	}

	assert.Equal(t, "6,4,2,1,3,5", DumpList64(sa, l))
	getItem = l.GetItems()

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		l.RemoveItem(sa, i)
	}

	assert.Equal(t, "", DumpList64(sa, l))
}

func TestList64SetHeadAndTail(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize64)
	l := new(list.List64).Init()

	for n := 1; n <= 6; n++ {
		l.AppendItem(sa, int64(n*list.ItemSize64))
	}

	assert.Equal(t, "1,2,3,4,5,6", DumpList64(sa, l))
	l.SetHead(sa, 3*list.ItemSize64)
	assert.Equal(t, "3,4,5,6,1,2", DumpList64(sa, l))
	l.SetTail(sa, 5*list.ItemSize64)
	assert.Equal(t, "6,1,2,3,4,5", DumpList64(sa, l))

	for n := 1; n <= 5; n++ {
		l.RemoveItem(sa, int64(n*list.ItemSize64))
	}

	assert.Equal(t, "6", DumpList64(sa, l))
	l.SetHead(sa, 6*list.ItemSize64)
	assert.Equal(t, "6", DumpList64(sa, l))
	l.SetTail(sa, 6*list.ItemSize64)
	assert.Equal(t, "6", DumpList64(sa, l))
}

func TestList64StoreAndLoad(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize64)
	l := new(list.List64).Init()

	for n := 1; n <= 6; n++ {
		l.AppendItem(sa, int64(n*list.ItemSize64))
	}

	assert.Equal(t, "1,2,3,4,5,6", DumpList64(sa, l))
	b := [list.Size64]byte{}
	l.Store(b[:])
	l.Load(b[:])
	assert.Equal(t, "1,2,3,4,5,6", DumpList64(sa, l))
}

func TestList64Flags(t *testing.T) {
	sa := make([]byte, 7*list.ItemSize64)
	l := new(list.List64).Init()

	for n := 1; n <= 6; n++ {
		l.AppendItem(sa, int64(n*list.ItemSize64))
		list.SetItem64Flags(sa, int64(n*list.ItemSize64), int8(n%4))
	}

	assert.Equal(t, "1,2,3,4,5,6", DumpList64(sa, l))
	getItem := l.GetItems()

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		assert.Equal(t, int8(i/list.ItemSize64%4), list.Item64Flags(sa, i))
	}
}

func DumpList64(sa []byte, l *list.List64) string {
	getItem := l.GetItems()
	text := ""

	for i, ok := getItem(sa); ok; i, ok = getItem(sa) {
		text += fmt.Sprintf("%d,", int(i/list.ItemSize64))
	}

	if text != "" {
		text = text[:len(text)-1]
	}

	return text
}
