// Package list implements a doubly-linked list.
package list

// List represents a doubly-linked list.
type List struct {
	nil item
}

// Init initializes the doubly-linked list and returns it.
func (l *List) Init() *List {
	l.nil = item{&l.nil, &l.nil, -1}
	return l
}

// AppendValue appends the given value to the doubly-linked list.
func (l *List) AppendValue(value int64) {
	item := item{Value: value}
	item.Insert(l.nil.Prev, &l.nil)
}

// PrependValue prepend the given value to the doubly-linked list.
func (l *List) PrependValue(value int64) {
	item := item{Value: value}
	item.Insert(&l.nil, l.nil.Next)
}

// GetValues returns an iteration function to iterate over
// all values in the doubly-linked list.
func (l *List) GetValues() func() (int64, bool) {
	item := l.nil.Next

	return func() (int64, bool) {
		if item == &l.nil {
			return 0, false
		}

		value := item.Value
		item = item.Next
		return value, true
	}
}

// GetPendingValues returns an iteration function to iterate over
// all values pending in the doubly-linked list.
func (l *List) GetPendingValues() func() (PendingValue, bool) {
	lastItem := l.nil.Prev
	item := l.nil.Next
	itemNext := item.Next

	return func() (PendingValue, bool) {
		if item == &l.nil {
			return PendingValue{}, false
		}

		pendingValue := pendingValue{&l.nil, item}

		if item == lastItem {
			item = &l.nil
		} else {
			item = itemNext
			itemNext = item.Next
		}

		return PendingValue(pendingValue), true
	}
}

// PendingValue represents a value pending in a doubly-linked list.
type PendingValue pendingValue

// MoveToBack move the value to the end of a doubly-linked list.
func (pv *PendingValue) MoveToBack() {
	pv.Item.Remove()
	pv.Item.Insert(pv.Nil.Prev, pv.Nil)
}

// MoveToFront move the value to the beginning of a doubly-linked list.
func (pv *PendingValue) MoveToFront() {
	pv.Item.Remove()
	pv.Item.Insert(pv.Nil, pv.Nil.Next)
}

// Delete deletes the value.
func (pv *PendingValue) Delete() {
	pv.Item.Remove()
}

// Value returns the value.
func (pv *PendingValue) Value() int64 {
	return pv.Item.Value
}

type item struct {
	Prev, Next *item
	Value      int64
}

func (i *item) Insert(prev *item, next *item) {
	i.Prev = prev
	prev.Next = i
	i.Next = next
	next.Prev = i
}

func (i *item) Remove() {
	i.Prev.Next = i.Next
	i.Next.Prev = i.Prev
}

type pendingValue struct {
	Nil  *item
	Item *item
}
