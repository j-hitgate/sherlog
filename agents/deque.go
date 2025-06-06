package agents

import "errors"

type Node[T any] struct {
	Value T
	Next  *Node[T]
	Prev  *Node[T]
}

type Deque[T any] struct {
	first *Node[T]
	last  *Node[T]
	len   int
}

func NewDeque[T any](values ...T) *Deque[T] {
	d := &Deque[T]{}

	if len(values) > 0 {
		d.Append(values...)
	}

	return d
}

func (d *Deque[T]) Len() int {
	return d.len
}

// Append and Pop

func (d *Deque[T]) Append(values ...T) {
	for _, val := range values {
		d.len++
		node := &Node[T]{Value: val}

		if d.len == 1 {
			d.first, d.last = node, node
			continue
		}

		node.Prev = d.last
		d.last.Next = node
		d.last = node
	}
}

func (d *Deque[T]) AppendLeft(values ...T) {
	for _, val := range values {
		d.len++
		node := &Node[T]{Value: val}

		if d.len == 1 {
			d.first, d.last = node, node
			continue
		}

		node.Next = d.first
		d.first.Prev = node
		d.first = node
	}
}

func (d *Deque[T]) Pop() (T, error) {
	if d.len == 0 {
		var t T
		return t, errors.New("empty deque")
	}

	d.len--
	node := d.last

	if d.len == 0 {
		d.first, d.last = nil, nil
		return node.Value, nil
	}

	d.last = node.Prev
	d.last.Next = nil

	return node.Value, nil
}

func (d *Deque[T]) PopLeft() (T, error) {
	if d.len == 0 {
		var t T
		return t, errors.New("empty deque")
	}

	d.len--
	node := d.first

	if d.len == 0 {
		d.first, d.last = nil, nil
		return node.Value, nil
	}

	d.first = node.Next
	d.first.Prev = nil

	return node.Value, nil
}

func (d *Deque[T]) Clear() {
	d.first, d.last, d.len = nil, nil, 0
}

// Getters

func (d *Deque[T]) Last() (T, error) {
	if d.len == 0 {
		var t T
		return t, errors.New("empty deque")
	}

	return d.last.Value, nil
}

func (d *Deque[T]) First() (T, error) {
	if d.len == 0 {
		var t T
		return t, errors.New("empty deque")
	}

	return d.first.Value, nil
}

func (d *Deque[T]) ToSlice() []T {
	arr := make([]T, d.len)
	node := d.first

	for i := range arr {
		arr[i] = node.Value
		node = node.Next
	}

	return arr
}
