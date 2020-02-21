package list_test

import (
	"testing"

	"github.com/roy2220/fsm/internal/list"
	"github.com/stretchr/testify/assert"
)

func TestListAddValue(t *testing.T) {
	l := new(list.List).Init()
	l.PrependValue(3)
	l.PrependValue(2)
	l.PrependValue(1)
	l.AppendValue(4)
	l.AppendValue(5)
	l.AppendValue(6)
	getValue := l.GetValues()
	n := int64(1)

	for v, ok := getValue(); ok; v, ok = getValue() {
		assert.Equal(t, n, v)
		n++
	}

	assert.Equal(t, int64(7), n)
}

func TestListDeleteValue(t *testing.T) {
	l := new(list.List).Init()

	for i := int64(0); i < 6; i++ {
		l.AppendValue(i + 1)
	}

	getPendingValue := l.GetPendingValues()

	for pv, ok := getPendingValue(); ok; pv, ok = getPendingValue() {
		if pv.Value()%2 == 0 {
			pv.Delete()
		}
	}

	getValue := l.GetValues()
	n := int64(1)

	for v, ok := getValue(); ok; v, ok = getValue() {
		assert.Equal(t, n, v)
		n += 2
	}

	assert.Equal(t, int64(7), n)
	getPendingValue = l.GetPendingValues()

	for pv, ok := getPendingValue(); ok; pv, ok = getPendingValue() {
		pv.Delete()
	}

	getValue = l.GetValues()
	_, ok := getValue()
	assert.False(t, ok)
}

func TestListMoveValue(t *testing.T) {
	l := new(list.List).Init()

	for i := int64(0); i < 6; i++ {
		l.AppendValue(i + 1)
	}

	getPendingValue := l.GetPendingValues()

	for pv, ok := getPendingValue(); ok; pv, ok = getPendingValue() {
		if pv.Value() >= 4 {
			pv.MoveToFront()
		}
	}

	getValue := l.GetValues()
	s := []int64{6, 5, 4, 1, 2, 3}
	i := 0

	for v, ok := getValue(); ok; v, ok = getValue() {
		assert.Equal(t, s[i], v)
		i++
	}

	assert.Equal(t, len(s), i)
	getPendingValue = l.GetPendingValues()
	i = 0
	f := true

	for pv, ok := getPendingValue(); ok; pv, ok = getPendingValue() {
		assert.Equal(t, s[i], pv.Value())
		i++

		if f && pv.Value() >= 4 {
			pv.MoveToBack()
		} else {
			f = false
		}
	}

	assert.Equal(t, len(s), i)
	getValue = l.GetValues()
	s = []int64{1, 2, 3, 6, 5, 4}
	i = 0

	for v, ok := getValue(); ok; v, ok = getValue() {
		assert.Equal(t, s[i], v)
		i++
	}

	assert.Equal(t, len(s), i)
}
