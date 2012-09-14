// This program demonstrates creating a Stream that emits the elements in the
// power set of an arbitrarily large set. The Power function must use the
// functional.Deferred function to lazily create the Stream. Without
// functional.Deferred, Power would eagerly calculate the entire power
// set stream causing the computer to freeze if asked to compute the power set
// of a set with more than just a few elements.
package main

import (
  "fmt"
  "github.com/keep94/gofunctional/functional"
)

var kEmptySetOnly [](*[]int) = [](*[]int){&[]int{}}

// Power returns the power set of items as a Stream of []int. Returned
// Stream expects to emit to a slice with capacity equal to len(items)
func Power(items []int) functional.Stream {
  length := len(items)
  if length == 0 {
    return functional.NewStreamFromPtrs(kEmptySetOnly, intSliceCopier)
  }
  newItems := items[0:length-1]
  return functional.Concat(
      Power(newItems), 
      functional.Deferred(func() functional.Stream {
          return functional.Filter(
              appendFilterer(items[length-1]), Power(newItems))
      }))
}

// intSliceCopier copies the values in a source []int to a dest []int.
func intSliceCopier(src, dest interface{}) {
  p := src.(*[]int)
  q := dest.(*[]int)
  length := copy(*q, *p)
  *q = (*q)[0:length]
}

// appendFilterer adds a particular int to an existing set.
type appendFilterer int

func (a appendFilterer) Filter(ptr interface{}) bool {
  p := ptr.(*[]int)
  *p = append(*p, int(a))
  return true
}

func main() {
  orig := make([]int, 100)
  for i := range orig {
    orig[i] = i
  }

  // Return the 10000th up to the 10010th element of the power set of
  // {0, 1, .. 99}.
  // This entire power set would have 2^100 elements in it!
  s := functional.Slice(Power(orig), 10000, 10010)
  result := make([]int, len(orig))
  for s.Next(&result) {
    fmt.Println(result)
  }
}
