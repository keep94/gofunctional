package functional

import (
    "errors"
    "fmt"
    "strings"
    "testing"
)

var (
  scanError = errors.New("error scanning.")
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
  s = Map(m, Filter(f, s), new(int))
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
  AppendValues(Map(doubleInt32Int64, Map(m, s, new(int)), new(int32)), &results)
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
  s := Map(squareIntInt32, xrange(3, 6), new(int))
  s = Map(doubleInt32Int64, s, new(int32))
  _, mapInMap := s.(*mapStream).stream.(*mapStream)
  if mapInMap {
    t.Error("Got a map within a map.")
  }
}

func TestNestedMap(t *testing.T) {
  s := Map(squareIntInt32, xrange(3, 6), new(int))
  s = Map(doubleInt32Int64, s, new(int32))
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
  AppendPtrs(xrange(1, 3), &results, func() interface{} { return new(int)})
  if *results[0] != 1 || *results[1] != 2 {
    t.Error("Wrong slice of pointers returned")
  }
}

func TestAppendPtrsNilCreater(t *testing.T) {
  var results []*int
  AppendPtrs(xrange(1, 3), &results, nil)
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

func TestCycleValues(t *testing.T) {
  var results[] int
  AppendValues(Slice(CycleValues([]int {3, 5}), 0, 4), &results)
  if output := fmt.Sprintf("%v", results); output != "[3 5 3 5]"  {
    t.Errorf("Expected [3 5 3 5] got %v", output)
  }
}

func TestCycleValuesWithEmptySlice(t *testing.T) {
  s := CycleValues([]int {})
  var x int
  if s.Next(&x) {
    t.Error("Expected empty Stream.")
  }
}

func TestCyclePtrs(t *testing.T) {
  var results[] int
  AppendValues(Slice(CyclePtrs([]*int {ptrInt(3), ptrInt(5)}, nil), 0, 4), &results)
  if output := fmt.Sprintf("%v", results); output != "[3 5 3 5]"  {
    t.Errorf("Expected [3 5 3 5] got %v", output)
  }
}

func TestCyclePtrsWithCopy(t *testing.T) {
  var results[] int
  AppendValues(Slice(CyclePtrs([]*int {ptrInt(3), ptrInt(5)}, copyInt), 0, 4), &results)
  if output := fmt.Sprintf("%v", results); output != "[3 5 3 5]"  {
    t.Errorf("Expected [3 5 3 5] got %v", output)
  }
}

func TestNewStreamFromPtrs(t *testing.T) {
  var results[] int
  AppendValues(NewStreamFromPtrs([]*int {ptrInt(3), ptrInt(5)}, nil), &results)
  if output := fmt.Sprintf("%v", results); output != "[3 5]"  {
    t.Errorf("Expected [3 5] got %v", output)
  }
}

func TestNewStreamFromPtrsWithCopy(t *testing.T) {
  var results[] int
  AppendValues(NewStreamFromPtrs([]*int {ptrInt(3), ptrInt(5)}, copyInt), &results)
  if output := fmt.Sprintf("%v", results); output != "[3 5]"  {
    t.Errorf("Expected [3 5] got %v", output)
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

func TestReadLines(t *testing.T) {
  r := strings.NewReader("Now is\nthe time\nfor all good men.\n")
  s := ReadLines(r)
  var results[] string
  AppendValues(s, &results)
  if output := strings.Join(results,","); output != "Now is,the time,for all good men."  {
    t.Errorf("Expected 'Now is,the time,for all good men' got '%v'", output)
  }
}

func TestReadLinesLongLine(t *testing.T) {
  str := strings.Repeat("a", 4001) + strings.Repeat("b", 4001) + strings.Repeat("c", 4001) + "\n" + "foo"
  s := ReadLines(strings.NewReader(str))
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

func TestReadLinesLongLine2(t *testing.T) {
  str := strings.Repeat("a", 4001) + strings.Repeat("b", 4001) + strings.Repeat("c", 4001)
  s := ReadLines(strings.NewReader(str))
  var results[] string
  AppendValues(s, &results)
  if results[0] != str {
    t.Error("Long line failed.")
  }
  if len(results) != 1 {
    t.Error("Results wrong length")
  }
}

func TestReadRows(t *testing.T) {
  rows := &fakeRows{ids: []int {3, 4}, names: []string{"foo", "bar"}}
  s := ReadRows(rows)
  var results []intAndString
  AppendValues(s, &results)
  if output := fmt.Sprintf("%v", results); output != "[{3 foo} {4 bar}]"  {
    t.Errorf("Expected [{3 foo} {4 bar}] got %v", output)
  }
} 

func TestReadRowsEmpty(t *testing.T) {
  rows := &fakeRows{ids: []int {}, names: []string {}}
  s := ReadRows(rows)
  var results []intAndString
  AppendValues(s, &results)
  if output := fmt.Sprintf("%v", results); output != "[]"  {
    t.Errorf("Expected [] got %v", output)
  }
} 

func TestReadRowsError(t *testing.T) {
  rows := &fakeRowsError{}
  s := ReadRows(rows)
  var result intAndString
  func() {
    defer func() {
      if x := recover(); x != scanError {
        t.Errorf("Expected scanError got %v", x)
      }
    }()
    s.Next(&result)
    t.Error("Expected error reading rows.")
  }()
}

func TestPartitionValues(t *testing.T) {
  expectedValues := []string {"[0 1 2]", "[3 4 5]", "[6]"}
  s := xrange(0, 7)
  mySlice := make([]int, 3)
  s = PartitionValues(s)
  var i int
  for i = 0; s.Next(&mySlice); i++ {
    if output := fmt.Sprintf("%v", mySlice); output != expectedValues[i] {
      t.Errorf("Expected %v but got %v", expectedValues[i], output)
    }
  }
  if i != len(expectedValues) {
    t.Errorf("Next terminated too soon")
  }
}

func TestPartitionValuesEmpty(t *testing.T) {
  s := xrange(0, 0)
  mySlice := make([]int, 3)
  s = PartitionValues(s)
  if s.Next(&mySlice) {
    t.Error("Next should return false on an empty Stream.")
  }
}

func TestPartitionPtrs(t *testing.T) {
  expectedValues := [][]int {{0, 1, 2}, {3, 4, 5}, {6}}
  s := xrange(0, 7)
  mySlice := make([]*int, 3)
  InitSlicePtrs(&mySlice, nil)
  s = PartitionPtrs(s)
  var i int
  for i = 0; s.Next(&mySlice); i++ {
    if len(mySlice) != len(expectedValues[i]) {
      t.Errorf("Expected length %v but got %v", len(expectedValues[i]), len(mySlice))
    }
    for j := range mySlice {
      if *mySlice[j] != expectedValues[i][j] {
        t.Errorf("Expected %v but got %v", expectedValues[i][j], *mySlice[j])
      }
    }
  }
  if i != len(expectedValues) {
    t.Errorf("Next terminated too soon")
  }
}

func TestPartitionPtrsEmpty(t *testing.T) {
  s := xrange(0, 0)
  mySlice := make([]*int, 3)
  InitSlicePtrs(&mySlice, nil)
  s = PartitionValues(s)
  if s.Next(&mySlice) {
    t.Error("Next should return false on an empty Stream.")
  }
}

func TestInitSlicePtrs(t *testing.T) {
  mySlice := make([]*int, 3)
  InitSlicePtrs(&mySlice, func() interface{} { return new(int) })
  if *mySlice[0] != 0 || *mySlice[1] != 0 || *mySlice[2] != 0 {
    t.Error("InitSlicePtrs failed")
  }
}
  
func TestGroupBy(t *testing.T) {
  s := Slice(CountFrom(6, 6), 0, 4)
  k := func(x interface{}) interface{} {
    p := x.(*int)
    return *p / 10
  }
  expected := []groupByResult{{0, []int{6}}, {1, []int{12, 18}}, {2, []int{24}}}
  s = GroupBy(s, k, new(int), nil)
  var group *Group
  var i int
  for i = 0; s.Next(&group); i++ {
    if key := group.Key().(int); key != expected[i].key {
      t.Errorf("Expected key %v got %v", expected[i].key, key)
    }
    var results []int
    AppendValues(group, &results)
    expectedValues := fmt.Sprintf("%v", expected[i].values)
    if output := fmt.Sprintf("%v", results); output != expectedValues {
      t.Errorf("Expected values %v got %v", expectedValues, output)
    }
    if key := group.Key().(int); key != expected[i].key {
      t.Errorf("Expected key %v got %v", expected[i].key, key)
    }
  }
  if i != len(expected) {
    t.Errorf("Next terminated too soon")
  }
}

func TestGroupBySkipping(t *testing.T) {
  s := Slice(CountFrom(6, 6), 0, 4)
  k := func(x interface{}) interface{} {
    p := x.(*int)
    return *p / 10
  }
  expected := []int {0, 1, 2}
  s = GroupBy(s, k, new(int), nil)
  var group *Group
  var i int
  for i = 0; s.Next(&group); i++ {
    if key := group.Key().(int); key != expected[i] {
      t.Errorf("Expected key %v got %v", expected[i], key)
    }
  }
  if i != len(expected) {
    t.Errorf("Next terminated too soon")
  }
}

func TestFlatten(t *testing.T) {
  if result := getNthDigit(15); result != 2 {
    t.Errorf("Expected 2 got %v", result)
  }
  if result := getNthDigit(300); result != 6 {
    t.Errorf("Expected 6 got %v", result)
  }
  if result := getNthDigit(188); result != 9 {
    t.Errorf("Expected 9 got %v", result)
  }
}

func TestFlattenWithEmptyStreams(t *testing.T) {
  first := NewStreamFromValues([]int{})
  second := NewStreamFromValues([]int{2})
  third := NewStreamFromValues([]int{})
  s := NewStreamFromValues([]Stream{first, second, third})
  var results []int
  AppendValues(Flatten(s), &results)
  if output := fmt.Sprintf("%v", results); output != "[2]" {
    t.Errorf("Expected [2] got %v", output)
  }
}

func TestDeferred(t *testing.T) {
  s := Deferred(func() Stream { return NewStreamFromValues([]int{2}) })
  var results []int
  AppendValues(s, &results)
  if output := fmt.Sprintf("%v", results); output != "[2]" {
    t.Errorf("Expected [2] got %v", output)
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
  c := Compose(g, f, func() interface{} { return new(int32)})
  c = Compose(h, c, func() interface{} { return new(int64)})
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

type intAndString struct {
  id int
  name string
}

func (t *intAndString) Ptrs() []interface{} {
  return []interface{}{&t.id, &t.name}
}

type fakeRows struct {
  ids []int
  names []string
  idx int
}

func (f *fakeRows) Next() bool {
  if f.idx == len(f.ids) || f.idx == len(f.names) {
    return false
  }
  f.idx++
  return true
}

func (f *fakeRows) Scan(args ...interface{}) error {
  p, q := args[0].(*int), args[1].(*string)
  *p = f.ids[f.idx - 1]
  *q = f.names[f.idx - 1]
  return nil
}

type fakeRowsError struct {}

func (f *fakeRowsError) Next() bool {
  return true
}

func (f *fakeRowsError) Scan(args ...interface{}) error {
  return scanError
}
  
type groupByResult struct {
  key int
  values []int
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

func copyInt(src, dest interface{}) {
  p := src.(*int)
  q := dest.(*int)
  *q = *p
}

// getNthDigit returns the nth digit in the sequence:
// 12345678910111213141516... getNthDigit(1) == 1.
func getNthDigit(x int) int {
  s := Slice(digitStream(), x - 1, -1)
  var result int
  s.Next(&result)
  return result
}

// digitStream returns a Stream of int = 1,2,3,4,5,6,7,8,9,1,0,1,1,...
func digitStream() Stream {
  return Flatten(Map(&intToDigitsMapper{}, Count(), new(int)))
}

// intToDigitsMapper converts an int into a Stream of int that emits its digits,
// most significant first.
type intToDigitsMapper struct {
  digits []int
}

// Map maps 123 -> {1, 2, 3}. Resulting Stream is valid until the next call
// to Map.
func (m *intToDigitsMapper) Map(srcPtr, destPtr interface{}) bool {
  x := *(srcPtr.(*int))
  result := destPtr.(*Stream)
  m.digits = m.digits[:0]
  for x > 0 {
    m.digits = append(m.digits, x % 10)
    x /= 10
  }
  for i := 0; i < len(m.digits) - i - 1; i++ {
    temp := m.digits[i]
    m.digits[i] = m.digits[len(m.digits) - i - 1]
    m.digits[len(m.digits) - i - 1] = temp
  }
  *result = NewStreamFromValues(m.digits)
  return true
}

func (m *intToDigitsMapper) Fast() Mapper {
  return m
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
