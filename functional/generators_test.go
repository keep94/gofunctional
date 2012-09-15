package functional

import (
    "fmt"
    "testing"
)

func TestNewInfiniteGenerator(t *testing.T) {
  g := NewGenerator(
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
  AppendValues(Slice(g, 0, 7), &results)
  if output := fmt.Sprintf("%v", results); output != "[0 1 1 2 3 5 8]"  {
    t.Errorf("Expected [0 1 1 2 3 5 8] got %v", output)
  }
  g.Close()
  if g.Next(new(int)) {
    t.Error("Next should return false after Close")
  }
}

func TestNewFiniteGenerator(t *testing.T) {

  g := NewGenerator(
      func(e Emitter) {
        values := []int{1, 2, 5}
        for _, x := range values {
          ptr := e.EmitPtr()
          if ptr == nil {
            break
          }
          *ptr.(*int) = x
        }
      })
  var results []int
  AppendValues(g, &results)
  if output := fmt.Sprintf("%v", results); output != "[1 2 5]" {
    t.Errorf("Expected [1 2 5] got %v", output)
  }
  g.Close()
}
