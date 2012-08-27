package odqueue

import (
    "fmt"
    "testing"
)

func TestQueue(t *testing.T) {
  q := NewQueue()
  p1 := q.End()
  p2 := p1
  q.Add(1).Add(2)
  if output := fmt.Sprintf("%v", toSlice(&p1)); output != "[1 2]" {
    t.Errorf("Expected [1 2] got %v", output)
  }
  q.Add(3)
  if output := fmt.Sprintf("%v", toSlice(&p1)); output != "[3]" {
    t.Errorf("Expected [3] got %v", output)
  }
  if output := fmt.Sprintf("%v", toSlice(&p2)); output != "[1 2 3]" {
    t.Errorf("Expected [1 2 3] got %v", output)
  }
}

func toSlice(p **Element) []interface{} {
  var result []interface{}
  for !(*p).IsEnd() {
    result = append(result, (*p).Value)
    *p = (*p).Next()
  }
  return result
}
