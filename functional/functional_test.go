package functional

import (
    "fmt"
    "strings"
    "testing"
)

func TestFilterAndMap(t *testing.T) {
  s := xrange(5, 15)
  f := NewFilterer(func(ptr interface{}) bool {
    p := ptr.(*int)
    return *p % 2 == 0
  })
  m := NewMapper(func(srcPtr interface{}, destPtr interface{}) bool {
    s := srcPtr.(*int)
    d := destPtr.(*int32)
    *d = int32((*s) * (*s))
    return true
  })
  s = Map(m, Filter(f, s), NewCreater(new(int)))
  var results []int32
  AppendValues(s, &results)
  if output := fmt.Sprintf("%v", results); output != "[36 64 100 144 196]"  {
    t.Errorf("Expected [36 64 100 144 196] got %v", output)
  }
}

func TestCombineFilterMap(t *testing.T) {
  s := xrange(5, 15)
  m := NewMapper(func(srcPtr interface{}, destPtr interface{}) bool {
    s := srcPtr.(*int)
    d := destPtr.(*int32)
    if *s % 2 != 0 {
      return false
    }
    *d = int32((*s) * (*s))
    return true
  })
  var results []int64
  AppendValues(Map(doubleInt32Int64, Map(m, s, NewCreater(new(int))), NewCreater(new(int32))), &results)
  if output := fmt.Sprintf("%v", results); output != "[72 128 200 288 392]"  {
    t.Errorf("Expected [64 128 200 288 392] got %v", output)
  }
}

func TestNoFilterInFilter(t *testing.T) {
  s := Filter(greaterThan(5), Filter(lessThan(8), xrange(0, 10)))
  _, filterInFilter := s.(*filterStream).stream.(*filterStream)
  if filterInFilter {
    t.Error("Got a filter within a filter.")
  }
}

func TestNestedFilter(t *testing.T) {
  s := Filter(greaterThan(5), Filter(lessThan(8), xrange(0, 10)))
  var results []int
  AppendValues(s, &results)
  if output := fmt.Sprintf("%v", results); output != "[6 7]"  {
    t.Errorf("Expected [6 7] got %v", output)
  }
}

func TestNoMapInMap(t *testing.T) {
  s := Map(squareIntInt32, xrange(3, 6), NewCreater(new(int)))
  s = Map(doubleInt32Int64, s, NewCreater(new(int32)))
  _, mapInMap := s.(*mapStream).stream.(*mapStream)
  if mapInMap {
    t.Error("Got a map within a map.")
  }
}

func TestNestedMap(t *testing.T) {
  s := Map(squareIntInt32, xrange(3, 6), NewCreater(new(int)))
  s = Map(doubleInt32Int64, s, NewCreater(new(int32)))
  var results []int64
  AppendValues(s, &results)
  if output := fmt.Sprintf("%v", results); output != "[18 32 50]"  {
    t.Errorf("Expected [18 32 50] got %v", output)
  }
}

func TestSliceNoEnd(t *testing.T) {
  s := xrange(5, 13)
  var results []int
  AppendValues(Slice(s, 5, -1), &results)
  if output := fmt.Sprintf("%v", results); output != "[10 11 12]"  {
    t.Errorf("Expected [10 11 12] got %v", output)
  }
}

func TestSliceWithEnd(t *testing.T) {
  s := xrange(5, 13)
  var results []int
  AppendValues(Slice(s, 2, 4), &results)
  if output := fmt.Sprintf("%v", results); output != "[7 8]"  {
    t.Errorf("Expected [7 8] got %v", output)
  }
}

func TestSliceWithEnd2(t *testing.T) {
  s := xrange(5, 13)
  var results []int
  AppendValues(Slice(s, 0, 2), &results)
  if output := fmt.Sprintf("%v", results); output != "[5 6]"  {
    t.Errorf("Expected [5 6] got %v", output)
  }
}

func TestZeroSlice(t *testing.T) {
  s := xrange(5, 13)
  var results []int
  AppendValues(Slice(s, 2, 2), &results)
  if output := fmt.Sprintf("%v", results); output != "[]"  {
    t.Errorf("Expected [] got %v", output)
  }
  var x int
  s.Next(&x)
  if x != 7 {
    t.Error("Slice advanced Stream too far.")
  }
}

func TestSliceStartTooBig(t *testing.T) {
  s := xrange(5, 13)
  var results []int
  AppendValues(Slice(s, 20, 30), &results)
  if output := fmt.Sprintf("%v", results); output != "[]"  {
    t.Errorf("Expected [] got %v", output)
  }
}

func TestSliceEndTooBig(t *testing.T) {
  s := xrange(5, 13)
  var results []int
  AppendValues(Slice(s, 7, 10), &results)
  if output := fmt.Sprintf("%v", results); output != "[12]"  {
    t.Errorf("Expected [12] got %v", output)
  }
}

func TestSliceStartBiggerThanEnd(t *testing.T) {
  s := xrange(5, 13)
  var results []int
  AppendValues(Slice(s, 4, 3), &results)
  if output := fmt.Sprintf("%v", results); output != "[]"  {
    t.Errorf("Expected [] got %v", output)
  }
}

func TestAppendPtrs(t *testing.T) {
  var results []*int
  AppendPtrs(xrange(1, 3), &results, NewCreater(new(int)))
  if *results[0] != 1 || *results[1] != 2 {
    t.Error("Wrong slice of pointers returned")
  }
}

func TestCountFrom(t *testing.T) {
  var results []int
  AppendValues(Slice(CountFrom(5, 2), 1, 3), &results)
  if output := fmt.Sprintf("%v", results); output != "[7 9]"  {
    t.Errorf("Expected [7 9] got %v", output)
  }
}

func TestConcat(t *testing.T) {
  var results[]int
  x := xrange(5, 8)
  y := xrange(3, 3)
  z := xrange(9, 11)
  AppendValues(Concat(x, y, z), &results)
  if output := fmt.Sprintf("%v", results); output != "[5 6 7 9 10]"  {
    t.Errorf("Expected [5 6 7 9 10] got %v", output)
  }
}

func TestConcat2(t *testing.T) {
  var results []int
  x := xrange(2, 2)
  y := xrange(7, 9)
  z := xrange(5, 5)
  AppendValues(Concat(x, y, z), &results)
  if output := fmt.Sprintf("%v", results); output != "[7 8]"  {
    t.Errorf("Expected [7 8] got %v", output)
  }
}

func TestConcatEmpty(t *testing.T) {
  var results []int
  AppendValues(Concat(), &results)
  if output := fmt.Sprintf("%v", results); output != "[]"  {
    t.Errorf("Expected [] got %v", output)
  }
}

func TestConcatAllEmptyStreams(t *testing.T) {
  var results []int
  x := xrange(0, 0)
  y := xrange(0, 0)
  AppendValues(Concat(x, y), &results)
  if output := fmt.Sprintf("%v", results); output != "[]"  {
    t.Errorf("Expected [] got %v", output)
  }
}

func TestJoin(t *testing.T) {
  var results []pair
  c := Count()
  s := Slice(CountFrom(0, 100), 0, 3)
  AppendValues(Join(s, c), &results)
  if output := fmt.Sprintf("%v", results); output != "[{0 0} {100 1} {200 2}]"  {
    t.Errorf("Expected [{0 0} {100 1} {200 2}] got %v", output)
  }
  var x int
  c.Next(&x)
  if x != 3 {
    t.Error("Join advanced second iterator too far.")
  }
}

func TestCycle(t *testing.T) {
  var results[] int
  AppendValues(Slice(Cycle([]int {3, 5}), 0, 4), &results)
  if output := fmt.Sprintf("%v", results); output != "[3 5 3 5]"  {
    t.Errorf("Expected [3 5 3 5] got %v", output)
  }
}

func TestDropWhileTakeWhile(t *testing.T) {
  var results[] int
  s := TakeWhile(lessThan(10), DropWhile(lessThan(7), Count()))
  AppendValues(s, &results)
  if output := fmt.Sprintf("%v", results); output != "[7 8 9]"  {
    t.Errorf("Expected [7 8 9] got %v", output)
  }
}

func TestDropWhileEmpty(t *testing.T) {
  var results[] int
  s := DropWhile(lessThan(7), xrange(0, 7))
  AppendValues(s, &results)
  if output := fmt.Sprintf("%v", results); output != "[]"  {
    t.Errorf("Expected [] got %v", output)
  }
}

func TestTakeWhileEmpty(t *testing.T) {
  var results[] int
  s := TakeWhile(lessThan(0), xrange(0, 7))
  AppendValues(s, &results)
  if output := fmt.Sprintf("%v", results); output != "[]"  {
    t.Errorf("Expected [] got %v", output)
  }
}

func TestDropWhileFull(t *testing.T) {
  var results[] int
  s := DropWhile(lessThan(0), xrange(0, 3))
  AppendValues(s, &results)
  if output := fmt.Sprintf("%v", results); output != "[0 1 2]"  {
    t.Errorf("Expected [0 1 2] got %v", output)
  }
}

func TestTakeWhileFull(t *testing.T) {
  var results[] int
  s := TakeWhile(lessThan(3), xrange(0, 3))
  AppendValues(s, &results)
  if output := fmt.Sprintf("%v", results); output != "[0 1 2]"  {
    t.Errorf("Expected [0 1 2] got %v", output)
  }
}

func TestLines(t *testing.T) {
  r := strings.NewReader("Now is\nthe time\nfor all good men.\n")
  s := Lines(r)
  var results[] string
  AppendValues(s, &results)
  if output := strings.Join(results,","); output != "Now is,the time,for all good men."  {
    t.Errorf("Expected 'Now is,the time,for all good men' got '%v'", output)
  }
}

func TestLinesLongLine(t *testing.T) {
  str := strings.Repeat("a", 4001) + strings.Repeat("b", 4001) + strings.Repeat("c", 4001) + "\n" + "foo"
  s := Lines(strings.NewReader(str))
  var results[] string
  AppendValues(s, &results)
  if results[0] != str[0:12003] {
    t.Error("Long line failed.")
  }
  if results[1] != "foo" {
    t.Error("Short line failed")
  }
  if len(results) != 2 {
    t.Error("Results wrong length")
  }
}

func TestLineLinesLongLine2(t *testing.T) {
  str := strings.Repeat("a", 4001) + strings.Repeat("b", 4001) + strings.Repeat("c", 4001)
  s := Lines(strings.NewReader(str))
  var results[] string
  AppendValues(s, &results)
  if results[0] != str {
    t.Error("Long line failed.")
  }
  if len(results) != 1 {
    t.Error("Results wrong length")
  }
}

func TestAny(t *testing.T) {
  a := Any(equal(1), equal(2))
  b := Any()
  c := Any(equal(3))
  d := equal(4)
  e := Any(a, b, c, d)
  for i := 1; i <= 4; i++ {
    if !e.Filter(ptrInt(i)) {
      t.Error("Call to Any failed")
    }
  }
  if e.Filter(ptrInt(0)) {
    t.Error("Call to Any failed")
  }
  if x := len(e.(orFilterer)); x != 4 {
    t.Errorf("Expected length of or filter to be 4, got %v", x)
  }
}

func TestAll(t *testing.T) {
  a := All(notEqual(1), notEqual(2))
  b := All()
  c := All(notEqual(3))
  d := notEqual(4)
  e := All(a, b, c, d)
  for i := 1; i <= 4; i++ {
    if e.Filter(ptrInt(i)) {
      t.Error("Call to All failed")
    }
  }
  if !e.Filter(ptrInt(0)) {
    t.Error("Call to All failed")
  }
  if x := len(e.(andFilterer)); x != 4 {
    t.Errorf("Expected length of and filter to be 4, got %v", x)
  }
}

func TestAllAnyComposition(t *testing.T) {
  a := All(
    Any(equal(1), equal(2), equal(3)),
    Any(equal(4)))
  if x := len(a.(andFilterer)); x != 2 {
    t.Errorf("Expected length of and filter to be 2, got %v", x)
  }
}

func TestAnyAllComposition(t *testing.T) {
  a := Any(
    All(equal(1), equal(2), equal(3)),
    All(equal(4)))
  if x := len(a.(orFilterer)); x != 2 {
    t.Errorf("Expected length of or filter to be 2, got %v", x)
  }
}

func TestEmptyAny(t *testing.T) {
  a := Any()
  if a.Filter(ptrInt(0)) {
    t.Error("Empty Any failed.")
  }
}
  
func TestEmptyAll(t *testing.T) {
  a := All()
  if !a.Filter(ptrInt(0)) {
    t.Error("Empty All failed.")
  }
}

func TestCompose(t *testing.T) {
  f := squareIntInt32
  g := doubleInt32Int64
  h := int64Plus1
  var i32 int32
  var i64 int64
  c := Compose(g, f, NewCreater(&i32))
  c = Compose(h, c, NewCreater(&i64))
  if x := len(c.(*compositeMapper).mappers); x != 3 {
    t.Error("Composition of composite mapper wrong.")
  }
  var result int64
  if !c.Map(ptrInt(5), &result) {
    t.Error("Map returns false instead of true.")
  }
  if result != 51 {
    t.Error("Map returned wrong value.")
  }
  if i32 != 0 || i64 != 0 {
    t.Error("Mapper not thread safe.")
  }
}  

type pair struct {
  x int
  y int
}

func (p *pair) Ptrs() []interface{} {
  return []interface{}{&p.x, &p.y}
}


func xrange(start, end int) Stream {
  return Slice(Count(), start, end)
}

func lessThan(x int) Filterer {
  return NewFilterer(func(ptr interface{}) bool {
    p := ptr.(*int)
    return *p < x
  })
}

func greaterThan(x int) Filterer {
  return NewFilterer(func(ptr interface{}) bool {
    p := ptr.(*int)
    return *p > x
  })
}

func notEqual(x int) Filterer {
  return NewFilterer(func(ptr interface{}) bool {
    p := ptr.(*int)
    return *p != x
  })
}

func equal(x int) Filterer {
  return NewFilterer(func(ptr interface{}) bool {
    p := ptr.(*int)
    return *p == x
  })
}
  
func ptrInt(x int) *int {
  return &x
}

var squareIntInt32 Mapper = NewMapper(
  func (srcPtr interface{}, destPtr interface{}) bool {
    p := srcPtr.(*int)
    q := destPtr.(*int32)
    *q = int32(*p) * int32(*p)
    return true
  })

var doubleInt32Int64 Mapper = NewMapper(
  func (srcPtr interface{}, destPtr interface{}) bool {
    p := srcPtr.(*int32)
    q := destPtr.(*int64)
    *q = 2 * int64(*p)
    return true
  })

var int64Plus1 Mapper = NewMapper(
  func (srcPtr interface{}, destPtr interface{}) bool {
    p := srcPtr.(*int64)
    q := destPtr.(*int64)
    *q = (*p) + 1
    return true
  })

