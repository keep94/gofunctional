// Package odqueue implements an on-demand queue in that elements remain
// accessible in the queue until no one is interested in consuming them.
// It works like any other FIFO queue except that each client keeps
// its current position by holding onto an *Element pointer. A client
// gets to the next element in the queue by calling Next() on its *Element
// pointer which returns a new *Element pointer. Any elements
// that appear in the queue before all the *Element pointers that clients
// hold are in-accessible and will eventually be GCed.
package odqueue

// NewQueue creates and returns a new Queue containing only an end element.
func NewQueue() *Queue {
  n := newElement()
  return &Queue{n}
}

// Element represents an element in the queue.
type Element struct {
  // Value is the value stored in the queue element.
  Value interface{}
  next *Element
}

// Next returns the next element in the queue. Calling Next on an end element
// returns the same end element.
func (e *Element) Next() *Element {
  return e.next
}

// IsEnd returns true if this element marks the end of the queue.
func (e *Element) IsEnd() bool {
  return e == e.next
}

type Queue struct {
  // Element is the end of the queue.
  e *Element
}

// Add stores x in the end element of this Queue and appends a new
// end element.  Add returns its receiver for chaining.
func (q *Queue) Add(x interface{}) *Queue {
  q.e.Value = x
  n := newElement()
  q.e.next = n
  q.e = n
  return q
}

// End returns the end of this queue. Calling IsEnd on returned element
// returns true.
func (q *Queue) End() *Element {
  return q.e
}

func newElement() *Element {
  result := &Element{}
  result.next = result
  return result
}
  
