package functional

import (
    "fmt"
    "testing"
)

func TestNormal(t *testing.T) {
  s := Slice(Count(), 0, 5)
  ec := newEvenNumberConsumer()
  oc := newOddNumberConsumer()
  MultiConsume(s, new(int), nil, ec, oc)
  if output := fmt.Sprintf("%v", ec.results); output != "[0 2 4]" {
    t.Errorf("Expected [0 2 4] got %v", output)
  }
  if output := fmt.Sprintf("%v", oc.results); output != "[1 3]" {
    t.Errorf("Expected [1 3] got %v", output)
  }
}

func TestConsumersEndEarly(t *testing.T) {
  s := Count()
  nc := ModifyConsumerStream(newEvenNumberConsumer(), func(s Stream) Stream {
    return NilStream()
  })
  first5 := func(s Stream) Stream {
    return Slice(s, 0, 5)
  }
  ec := newEvenNumberConsumer()
  oc := newOddNumberConsumer()
  MultiConsume(
      s,
      new(int),
      nil,
      nc,
      ModifyConsumerStream(ec, first5),
      ModifyConsumerStream(oc, first5))
  if output := fmt.Sprintf("%v", ec.results); output != "[0 2 4]" {
    t.Errorf("Expected [0 2 4] got %v", output)
  }
  if output := fmt.Sprintf("%v", oc.results); output != "[1 3]" {
    t.Errorf("Expected [1 3] got %v", output)
  }
  var result int
  s.Next(&result)
  if result != 5 {
    t.Errorf("Expected 5 got %v", result)
  }
}

func TestNoConsumers(t *testing.T) {
  s := CountFrom(7, 1)
  MultiConsume(s, new(int), nil)
  var result int
  if !s.Next(&result) || result != 7 {
    t.Errorf("Expected 7 got %v", result)
  }
}

func TestReadPastEnd(t *testing.T) {
  s := Slice(Count(), 0, 5)
  rc := &readPastEndConsumer{}
  MultiConsume(s, new(int), nil, rc)
  if !rc.completed {
    t.Error("MultiConsume returned before child consumers completed.")
  }
}

type filterConsumer struct {
  f Filterer
  results []int
}

func (fc *filterConsumer) Consume(s Stream) {
  AppendValues(Filter(fc.f, s), &fc.results)
}

type readPastEndConsumer struct {
  completed bool
}

func (c *readPastEndConsumer) Consume(s Stream) {
  var x int
  for s.Next(&x) {
  }
  for i := 0; i < 10; i++ {
    s.Next(&x)
  }
  c.completed = true
}

func newEvenNumberConsumer() *filterConsumer {
  return &filterConsumer{f: NewFilterer(func(ptr interface{}) bool {
    p := ptr.(*int)
    return *p % 2 == 0
  })}
}

func newOddNumberConsumer() *filterConsumer {
  return &filterConsumer{f: NewFilterer(func(ptr interface{}) bool {
    p := ptr.(*int)
    return *p % 2 == 1
  })}
}
