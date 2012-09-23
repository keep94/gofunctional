package functional

import (
    "fmt"
    "testing"
)

func TestNewInfiniteGenerator(t *testing.T) {

  var finished bool
  // fibonacci
  fib := NewGenerator(
      func(e Emitter) {
        a := 0
        b := 1
        for ptr := e.EmitPtr(); ptr != nil; ptr = e.EmitPtr() {
          p := ptr.(*int)
          *p = a
          a, b = b, a + b
        }
        finished = true
      })
  var results []int
  first7Fibs := StreamToGenerator(Slice(fib, 0, 7), fib)
  AppendValues(first7Fibs, &results)
  if output := fmt.Sprintf("%v", results); output != "[0 1 1 2 3 5 8]"  {
    t.Errorf("Expected [0 1 1 2 3 5 8] got %v", output)
  }
  first7Fibs.Close()

  if !finished {
    t.Error("Generating function should complete on close.")
  }

  // Closing first7Fibs should also close underlying fib generator
  if fib.Next(new(int)) {
    t.Error("fib generator should be closed")
  }
}

func TestNewFiniteGenerator(t *testing.T) {
  g := NewGenerator(
      func(e Emitter) {
        values := []int{1, 2, 5}
        ptr := e.EmitPtr()
        for i := range values {
          if ptr == nil {
            break
          }
          *ptr.(*int) = values[i]
          ptr = e.EmitPtr()
        }
      })
  var results []int
  AppendValues(g, &results)
  if output := fmt.Sprintf("%v", results); output != "[1 2 5]" {
    t.Errorf("Expected [1 2 5] got %v", output)
  }
  if g.Next(new(int)) {
    t.Error("Next on generator should repeatedly return false after all value have been emitted. Got true")
  }
  g.Close()
}

func TestEmptyGenerator(t *testing.T) {
  g := NewGenerator(func (e Emitter) {})
  if g.Next(new(int)) {
    t.Error("Next should return false on empty generator, got true")
  }
  if g.Next(new(int)) {
    t.Error("Next on empty generator should repeatedly return false. Got true")
  }
  g.Close()
}
