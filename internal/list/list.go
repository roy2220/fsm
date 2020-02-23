// Package list implements a doubly-linked list.
package list

import "encoding/binary"

const (
	// Size is the size of a doubly-linked list.
	Size = 16

	// ItemSize is the size of a doubly-linked list item.
	ItemSize = 16
)

// List represents a doubly-linked list.
type List struct {
	tail, head item
}

// Init initializes the doubly-linked list and returns it.
func (l *List) Init() *List {
	l.Clear()
	return l
}

// AppendItem appends the given item to the doubly-linked list.
func (l *List) AppendItem(spaceAccessor []byte, rawItem int64) {
	item := item(rawItem)

	if l.IsEmpty() {
		item.SetPrev(spaceAccessor, item)
		item.SetNext(spaceAccessor, item)
		l.tail = item
		l.head = item
	} else {
		item.Insert(spaceAccessor, l.tail, l.head)
		l.tail = item
	}
}

// PrependItem prepends the given item to the doubly-linked list.
func (l *List) PrependItem(spaceAccessor []byte, rawItem int64) {
	item := item(rawItem)

	if l.IsEmpty() {
		item.SetPrev(spaceAccessor, item)
		item.SetNext(spaceAccessor, item)
		l.tail = item
		l.head = item
	} else {
		item.Insert(spaceAccessor, l.tail, l.head)
		l.head = item
	}
}

// RemoveItem removes the given item from the doubly-linked list.
func (l *List) RemoveItem(spaceAccessor []byte, rawItem int64) {
	if l.tail == l.head {
		l.Clear()
	} else {
		item := item(rawItem)
		itemPrev, itemNext := item.Remove(spaceAccessor)

		if item == l.tail {
			l.tail = itemPrev
		} else if item == l.head {
			l.head = itemNext
		}
	}
}

// SetTail sets the given item as the tail of the doubly-linked list.
func (l *List) SetTail(spaceAccessor []byte, rawItem int64) {
	item := item(rawItem)
	l.tail = item
	l.head = item.Next(spaceAccessor)
}

// SetHead sets the given item as the head of the doubly-linked list.
func (l *List) SetHead(spaceAccessor []byte, rawItem int64) {
	item := item(rawItem)
	l.head = item
	l.tail = item.Prev(spaceAccessor)
}

// GetItems returns an iteration function to iterate over
// all items in the doubly-linked list.
func (l *List) GetItems() func([]byte) (int64, bool) {
	lastItem, nextItem, item := l.tail, l.head, item(noItem)

	return func(spaceAccessor []byte) (int64, bool) {
		if item == lastItem {
			return 0, false
		}

		item = nextItem
		nextItem = nextItem.Next(spaceAccessor)
		return int64(item), true
	}
}

// Store stores the doubly-linked list to the given buffer.
func (l *List) Store(buffer *[Size]byte) {
	binary.BigEndian.PutUint64(buffer[:], ^uint64(l.tail))
	binary.BigEndian.PutUint64(buffer[8:], ^uint64(l.head))
}

// Load loads the doubly-linked list from the given data.
func (l *List) Load(data *[Size]byte) {
	l.tail = item(^binary.BigEndian.Uint64(data[:]))
	l.head = item(^binary.BigEndian.Uint64(data[8:]))
}

// Clear clears the doubly-linked list to empty.
func (l *List) Clear() {
	l.tail = noItem
	l.head = noItem
}

// IsEmpty returns a boolean value indicates whether the doubly-linked list is empty.
func (l *List) IsEmpty() bool {
	return l.tail == noItem
}

const noItem = -1

type item int64

func (i item) Insert(spaceAccessor []byte, prev, next item) {
	i.SetPrev(spaceAccessor, prev)
	prev.SetNext(spaceAccessor, i)
	i.SetNext(spaceAccessor, next)
	next.SetPrev(spaceAccessor, i)
}

func (i item) Remove(spaceAccessor []byte) (item, item) {
	prev, next := i.Prev(spaceAccessor), i.Next(spaceAccessor)
	prev.SetNext(spaceAccessor, next)
	next.SetPrev(spaceAccessor, prev)
	return prev, next
}

func (i item) SetPrev(spaceAccessor []byte, prev item) {
	binary.BigEndian.PutUint64(spaceAccessor[i:], uint64(prev))
}

func (i item) SetNext(spaceAccessor []byte, prev item) {
	binary.BigEndian.PutUint64(spaceAccessor[i+8:], uint64(prev))
}

func (i item) Prev(spaceAccessor []byte) item {
	return item(binary.BigEndian.Uint64(spaceAccessor[i:]))
}

func (i item) Next(spaceAccessor []byte) item {
	return item(binary.BigEndian.Uint64(spaceAccessor[i+8:]))
}
