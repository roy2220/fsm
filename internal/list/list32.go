// Package list implements a doubly-linked list.
package list

import "encoding/binary"

const (
	// Size32 is the size of a doubly-linked list.
	Size32 = 32 / 8 * 2

	// ItemSize32 is the size of a doubly-linked list item.
	ItemSize32 = 32 / 8 * 2
)

// List32 represents a doubly-linked list.
type List32 struct {
	tail, head item32
}

// Init initializes the doubly-linked list and returns it.
func (l *List32) Init() *List32 {
	l.Clear()
	return l
}

// AppendItem appends the given item to the doubly-linked list.
func (l *List32) AppendItem(spaceAccessor []byte, rawItem int32) {
	item := item32(rawItem)

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
func (l *List32) PrependItem(spaceAccessor []byte, rawItem int32) {
	item := item32(rawItem)

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
func (l *List32) InsertItemAfter(spaceAccessor []byte, rawItem, otherRawItem int32) {
	item := item32(rawItem)
	otherItem := item32(otherRawItem)
	item.Insert(spaceAccessor, otherItem, otherItem.Next(spaceAccessor))

	if otherItem == l.tail {
		l.tail = item
	}
}

// InsertItemBefore inserts the given item before the another one in the doubly-linked list.
func (l *List32) InsertItemBefore(spaceAccessor []byte, rawItem, otherRawItem int32) {
	item := item32(rawItem)
	otherItem := item32(otherRawItem)
	item.Insert(spaceAccessor, otherItem.Prev(spaceAccessor), otherItem)

	if otherItem == l.head {
		l.head = item
	}
}

// RemoveItem removes the given item from the doubly-linked list.
func (l *List32) RemoveItem(spaceAccessor []byte, rawItem int32) {
	if l.tail == l.head {
		l.Clear()
	} else {
		item := item32(rawItem)
		itemPrev, itemNext := item.Remove(spaceAccessor)

		if item == l.tail {
			l.tail = itemPrev
		} else if item == l.head {
			l.head = itemNext
		}
	}
}

// SetTail sets the given item as the tail of the doubly-linked list.
func (l *List32) SetTail(spaceAccessor []byte, rawItem int32) {
	item := item32(rawItem)
	l.tail = item
	l.head = item.Next(spaceAccessor)
}

// SetHead sets the given item as the head of the doubly-linked list.
func (l *List32) SetHead(spaceAccessor []byte, rawItem int32) {
	item := item32(rawItem)
	l.head = item
	l.tail = item.Prev(spaceAccessor)
}

// GetItems returns an iteration function to iterate over
// all items in the doubly-linked list.
func (l *List32) GetItems() func([]byte) (int32, bool) {
	lastItem, nextItem, item := l.tail, l.head, noItem32

	return func(spaceAccessor []byte) (int32, bool) {
		if item == lastItem {
			return 0, false
		}

		item = nextItem
		nextItem = nextItem.Next(spaceAccessor)
		return int32(item), true
	}
}

// Store stores the doubly-linked list to the given buffer.
func (l *List32) Store(buffer []byte) {
	_ = buffer[Size32-1]
	binary.BigEndian.PutUint32(buffer[:], ^uint32(l.tail))
	binary.BigEndian.PutUint32(buffer[32/8:], ^uint32(l.head))
}

// Load loads the doubly-linked list from the given data.
func (l *List32) Load(data []byte) {
	_ = data[Size32-1]
	l.tail = item32(^binary.BigEndian.Uint32(data[:]))
	l.head = item32(^binary.BigEndian.Uint32(data[32/8:]))
}

// Clear clears the doubly-linked list to empty.
func (l *List32) Clear() {
	l.tail = noItem32
	l.head = noItem32
}

// IsEmpty returns a boolean value indicates whether the doubly-linked list is empty.
func (l *List32) IsEmpty() bool {
	return l.tail == noItem32
}

// SetItem32Flags sets the flags(2) of the given item.
func SetItem32Flags(spaceAccessor []byte, rawItem int32, flags int8) {
	item32(rawItem).SetFlags(spaceAccessor, flags)
}

// Item32Flags returns the flags(2) of the given item.
func Item32Flags(spaceAccessor []byte, rawItem int32) int8 {
	return item32(rawItem).Flags(spaceAccessor)
}

// Item32Prev returns the predecessor of the given item.
func Item32Prev(spaceAccessor []byte, rawItem int32) int32 {
	return int32(item32(rawItem).Prev(spaceAccessor))
}

// Item32Next returns the successor of the given item.
func Item32Next(spaceAccessor []byte, rawItem int32) int32 {
	return int32(item32(rawItem).Next(spaceAccessor))
}

const noItem32 = item32(-1)

type item32 int32

func (i item32) Insert(spaceAccessor []byte, prev, next item32) {
	i.SetPrev(spaceAccessor, prev)
	prev.SetNext(spaceAccessor, i)
	i.SetNext(spaceAccessor, next)
	next.SetPrev(spaceAccessor, i)
}

func (i item32) Remove(spaceAccessor []byte) (item32, item32) {
	prev, next := i.Prev(spaceAccessor), i.Next(spaceAccessor)
	prev.SetNext(spaceAccessor, next)
	next.SetPrev(spaceAccessor, prev)
	return prev, next
}

func (i item32) SetFlags(spaceAccessor []byte, flags int8) {
	b1 := &spaceAccessor[i]
	*b1 = (*b1 &^ (1 << 7)) | (uint8(flags>>1) << 7)
	b2 := &spaceAccessor[i+32/8]
	*b2 = (*b2 &^ (1 << 7)) | (uint8(flags&1) << 7)
}

func (i item32) SetPrev(spaceAccessor []byte, prev item32) {
	buffer := spaceAccessor[i:]
	binary.BigEndian.PutUint32(buffer, uint32(prev)|(uint32(buffer[0])>>7<<(32-1)))
}

func (i item32) SetNext(spaceAccessor []byte, next item32) {
	buffer := spaceAccessor[i+32/8:]
	binary.BigEndian.PutUint32(buffer, uint32(next)|(uint32(buffer[0])>>7<<(32-1)))
}

func (i item32) Flags(spaceAccessor []byte) int8 {
	f1 := spaceAccessor[i] >> 7
	f2 := spaceAccessor[i+32/8] >> 7
	return int8((f1 << 1) | f2)
}

func (i item32) Prev(spaceAccessor []byte) item32 {
	return item32(binary.BigEndian.Uint32(spaceAccessor[i:]) &^ (1 << (32 - 1)))
}

func (i item32) Next(spaceAccessor []byte) item32 {
	return item32(binary.BigEndian.Uint32(spaceAccessor[i+32/8:]) &^ (1 << (32 - 1)))
}
