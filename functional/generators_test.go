package functional

import (
    "fmt"
    "testing"
)

func TestNewInfiniteGenerator(t *testing.T) {

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
      })
  var results []int
  first7Fibs := StreamToGenerator(Slice(fib, 0, 7), fib)
  AppendValues(first7Fibs, &results)
  if output := fmt.Sprintf("%v", results); output != "[0 1 1 2 3 5 8]"  {
    t.Errorf("Expected [0 1 1 2 3 5 8] got %v", output)
  }
  first7Fibs.Close()

  // Closing first7Fibs should also close underlying fib generator
  if fib.Next(new(int)) {
    t.Error("fib generator should be closed")
  }
}

func TestNewFiniteGenerator(t *testing.T) {
  g := NewGenerator(
      func(e Emitter) {
        values := []int{1, 2, 5}
        for i := range values {
          ptr := e.EmitPtr()
          if ptr == nil {
            break
          }
          *ptr.(*int) = values[i]
        }
      })
  var results []int
  AppendValues(g, &results)
  if output := fmt.Sprintf("%v", results); output != "[1 2 5]" {
    t.Errorf("Expected [1 2 5] got %v", output)
  }
  g.Close()
}

func TestEmptyGenerator(t *testing.T) {
  g := NewGenerator(func (e Emitter) {})
  if g.Next(new(int)) {
    t.Error("Next should return false on empty generator, got true")
  }
  g.Close()
}
