// Package list implements a doubly-linked list.
package list

import "encoding/binary"

const (
	// Size64 is the size of a doubly-linked list.
	Size64 = 64 / 8 * 2

	// ItemSize64 is the size of a doubly-linked list item.
	ItemSize64 = 64 / 8 * 2
)

// List64 represents a doubly-linked list.
type List64 struct {
	tail, head item64
}

// Init initializes the doubly-linked list and returns it.
func (l *List64) Init() *List64 {
	l.Clear()
	return l
}

// AppendItem appends the given item to the doubly-linked list.
func (l *List64) AppendItem(spaceAccessor []byte, rawItem int64) {
	item := item64(rawItem)

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
func (l *List64) PrependItem(spaceAccessor []byte, rawItem int64) {
	item := item64(rawItem)

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

// InsertItemAfter inserts the given item after the another one in the doubly-linked list.
func (l *List64) InsertItemAfter(spaceAccessor []byte, rawItem, otherRawItem int64) {
	item := item64(rawItem)
	otherItem := item64(otherRawItem)
	item.Insert(spaceAccessor, otherItem, otherItem.Next(spaceAccessor))

	if otherItem == l.tail {
		l.tail = item
	}
}

// InsertItemBefore inserts the given item before the another one in the doubly-linked list.
func (l *List64) InsertItemBefore(spaceAccessor []byte, rawItem, otherRawItem int64) {
	item := item64(rawItem)
	otherItem := item64(otherRawItem)
	item.Insert(spaceAccessor, otherItem.Prev(spaceAccessor), otherItem)

	if otherItem == l.head {
		l.head = item
	}
}

// RemoveItem removes the given item from the doubly-linked list.
func (l *List64) RemoveItem(spaceAccessor []byte, rawItem int64) {
	if l.tail == l.head {
		l.Clear()
	} else {
		item := item64(rawItem)
		itemPrev, itemNext := item.Remove(spaceAccessor)

		if item == l.tail {
			l.tail = itemPrev
		} else if item == l.head {
			l.head = itemNext
		}
	}
}

// SetTail sets the given item as the tail of the doubly-linked list.
func (l *List64) SetTail(spaceAccessor []byte, rawItem int64) {
	item := item64(rawItem)
	l.tail = item
	l.head = item.Next(spaceAccessor)
}

// SetHead sets the given item as the head of the doubly-linked list.
func (l *List64) SetHead(spaceAccessor []byte, rawItem int64) {
	item := item64(rawItem)
	l.head = item
	l.tail = item.Prev(spaceAccessor)
}

// GetItems returns an iteration function to iterate over
// all items in the doubly-linked list.
func (l *List64) GetItems() func([]byte) (int64, bool) {
	lastItem, nextItem, item := l.tail, l.head, noItem64

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
func (l *List64) Store(buffer []byte) {
	_ = buffer[Size64-1]
	binary.BigEndian.PutUint64(buffer[:], ^uint64(l.tail))
	binary.BigEndian.PutUint64(buffer[64/8:], ^uint64(l.head))
}

// Load loads the doubly-linked list from the given data.
func (l *List64) Load(data []byte) {
	_ = data[Size64-1]
	l.tail = item64(^binary.BigEndian.Uint64(data[:]))
	l.head = item64(^binary.BigEndian.Uint64(data[64/8:]))
}

// Clear clears the doubly-linked list to empty.
func (l *List64) Clear() {
	l.tail = noItem64
	l.head = noItem64
}

// IsEmpty returns a boolean value indicates whether the doubly-linked list is empty.
func (l *List64) IsEmpty() bool {
	return l.tail == noItem64
}

// SetItem64Flags sets the flags(2) of the given item.
func SetItem64Flags(spaceAccessor []byte, rawItem int64, flags int8) {
	item64(rawItem).SetFlags(spaceAccessor, flags)
}

// Item64Flags returns the flags(2) of the given item.
func Item64Flags(spaceAccessor []byte, rawItem int64) int8 {
	return item64(rawItem).Flags(spaceAccessor)
}

// Item64Prev returns the predecessor of the given item.
func Item64Prev(spaceAccessor []byte, rawItem int64) int64 {
	return int64(item64(rawItem).Prev(spaceAccessor))
}

// Item64Next returns the successor of the given item.
func Item64Next(spaceAccessor []byte, rawItem int64) int64 {
	return int64(item64(rawItem).Next(spaceAccessor))
}

const noItem64 = item64(-1)

type item64 int64

func (i item64) Insert(spaceAccessor []byte, prev, next item64) {
	i.SetPrev(spaceAccessor, prev)
	prev.SetNext(spaceAccessor, i)
	i.SetNext(spaceAccessor, next)
	next.SetPrev(spaceAccessor, i)
}

func (i item64) Remove(spaceAccessor []byte) (item64, item64) {
	prev, next := i.Prev(spaceAccessor), i.Next(spaceAccessor)
	prev.SetNext(spaceAccessor, next)
	next.SetPrev(spaceAccessor, prev)
	return prev, next
}

func (i item64) SetFlags(spaceAccessor []byte, flags int8) {
	b1 := &spaceAccessor[i]
	*b1 = (*b1 &^ (1 << 7)) | (uint8(flags>>1) << 7)
	b2 := &spaceAccessor[i+64/8]
	*b2 = (*b2 &^ (1 << 7)) | (uint8(flags&1) << 7)
}

func (i item64) SetPrev(spaceAccessor []byte, prev item64) {
	buffer := spaceAccessor[i:]
	binary.BigEndian.PutUint64(buffer, uint64(prev)|(uint64(buffer[0])>>7<<(64-1)))
}

func (i item64) SetNext(spaceAccessor []byte, next item64) {
	buffer := spaceAccessor[i+64/8:]
	binary.BigEndian.PutUint64(buffer, uint64(next)|(uint64(buffer[0])>>7<<(64-1)))
}

func (i item64) Flags(spaceAccessor []byte) int8 {
	f1 := spaceAccessor[i] >> 7
	f2 := spaceAccessor[i+64/8] >> 7
	return int8((f1 << 1) | f2)
}

func (i item64) Prev(spaceAccessor []byte) item64 {
	return item64(binary.BigEndian.Uint64(spaceAccessor[i:]) &^ (1 << (64 - 1)))
}

func (i item64) Next(spaceAccessor []byte) item64 {
	return item64(binary.BigEndian.Uint64(spaceAccessor[i+64/8:]) &^ (1 << (64 - 1)))
}
