// Package functional provides functional programming constructs.
package functional

import (
  "bufio"
  "fmt"
  "io"
  "reflect"
)

// Stream is a sequence of lazily produced values.
// Each call to Next() evaluates the next value in the stream.
type Stream interface {
  // Next evaluates the next value in this Stream.
  // If Next returns true, the next value is stored at ptr;
  // If Next returns false, then the end of the Stream has been reached,
  // and the value ptr points to is unspecified. ptr must be a
  // pointer type
  Next(ptr interface{}) bool
}

// Tuple represents a tuple of values that Join emits
type Tuple interface {
  // Ptrs returns a pointer to each field in the tuple.
  Ptrs() []interface{}
}

// Filterer filters values in a Stream.
type Filterer interface {
  // Filter returns true if value ptr points to should be included or false
  // otherwise. ptr must be a pointer type
  Filter(ptr interface{}) bool
}

// Mapper maps one value to another in a Stream.
type Mapper interface {
  // Map does the mapping. srcPtr points to the original value, and the 
  // mapped value is stored in destPtr. srcPtr and destPtr must be pointer
  // types. If Mapper returns false, then no mapped value stored at destPtr.
  Map(srcPtr interface{}, destPtr interface{}) bool
  // Returns a thread unsafe version of this Mapper. In particular any instances
  // needed to hold intermediate values are creaed in advance and reused in
  // each call to Map.
  Fast() Mapper
}

// Creater creates a new instance.
type Creater interface {
  // Create creates a new instance. Returns a pointer to the new instance.
  Create() interface{}
}

// Rows represents rows in a database table
type Rows interface {
  // Next advances to the next row. Next returns false if there is no next row.
  // Every call to Scan, even the first one, must be preceded by a call to Next.
  Next() bool
  // Reads the values out of the current row. args are pointer types.
  Scan(args ...interface{}) error
}

// Map applies f to s and returns the new Stream. If s is
// (x1, x2, x3, ...), Map returns the Stream (f(x1), f(x2), f(x3), ...).
// c creates the instance that s emits to.
func Map(f Mapper, s Stream, c Creater) Stream {
  ms, ok := s.(*mapStream)
  if ok {
    return &mapStream{Compose(f, ms.mapper, c).Fast(), ms.stream, ms.ptr}
  }
  return &mapStream{f.Fast(), s, c.Create()}
}

// Filter filters values from s and returns the resulting Stream.
// Filter f will return true for each value in the returned Stream.
func Filter(f Filterer, s Stream) Stream {
  fs, ok := s.(*filterStream)
  if ok {
    return &filterStream{All(fs.filterer, f), fs.stream}
  }
  return &filterStream{f, s}
}

// Count returns the infinite stream of integers beginning at 0.
func Count() Stream {
  return &count{0, 1}
}

// CountFrom returns the infinite stream of integers beginning at start
// and increasing by step.
func CountFrom(start, step int) Stream {
  return &count{start, step}
}

// Slice returns a Stream that will emit elements in s starting at index start
// and continuing to but not including index end. Indexes are 0 based. If end
// is negative, it means go to the end of s.
func Slice(s Stream, start int, end int) Stream {
  return &sliceStream{s, start, end, 0}
}

// Concat concatenates multiple Streams into one.
// If x = (x1, x2, ...) and y = (y1, y2, ...) then
// Concat(x, y) = (x1, x2, ..., y1, y2, ...)
func Concat(s ...Stream) Stream {
  return &concatStream{s, 0}
}

// Join uses multiple Streams to form a new Stream of Tuples.
// if x = (x1, x2, ..) and y = (y1, y2, ...) then
// Join(x, y) = ((x1, y1), (x2, y2), ...). 
// The Stream Join returns quits emitting whenever one of the input Streams
// runs out.
func Join(s ...Stream) Stream {
  return &joinStream{s, false}
}

// Cycle emits the elements in the given slice over and over again.
// Cycle([]int {3, 5}) = (3, 5, 3, 5, ...)
// The Stream Cycle returns owns the passed in slice, so clients should not
// later modify the slice passed to Cycle.
func Cycle(aSlice interface{}) Stream {
  sliceValue := reflect.ValueOf(aSlice)
  if sliceValue.Kind() != reflect.Slice {
    panic("Slice argument expected")
  }
  return &cycleStream{sliceValue, sliceValue.Len(), 0}
}

// TakeWhile returns a Stream that emits the values in s until f is false.
func TakeWhile(f Filterer, s Stream) Stream {
  return &takeStream{f, s, false}
}

// DropWhile returns a Stream that emits the values in s starting at the
// first value where f is false.
func DropWhile(f Filterer, s Stream) Stream {
  return &dropStream{f, s, false}
}

// ReadLines returns the lines of text in r as a Stream of string types.
// The returned lines are separated by either \n or \r\n. The emitted
// string types do not contain the end of line characters.
func ReadLines(r io.Reader) Stream {
  return lineStream{bufio.NewReader(r)}
}

// ReadRows returns the rows in a database table as a Stream of Tuple types.
func ReadRows(r Rows) Stream {
  return rowStream{r}
}

// AppendValues evaluates s and places each element in s
// in the slice that slicePtr points to.
// If s emits elements of type T, then the slice that slicePtr points
// to must be of type []T.  
func AppendValues(s Stream, slicePtr interface{}) {
  sliceValue := getSliceValue(slicePtr)
  sliceElementType := sliceValue.Type().Elem()
  sliceValue.Set(appendValues(s, sliceElementType, sliceValue))
}

// AppendPtrs evaluates s and places each element in s in the slice that
// slicePtr points to. If s emits elements of type T, then the slice that
// slicePtr points to must be of type []*T. c creates the instances that 
// the *T's point to.
func AppendPtrs(s Stream, slicePtr interface{}, c Creater) {
  sliceValue := getSliceValue(slicePtr)
  sliceElementType := sliceValue.Type().Elem()
  if sliceElementType.Kind() != reflect.Ptr {
    panic("slicePtr must point to a slice of pointers.")
  }
  sliceValue.Set(appendPtrs(s, c, sliceValue))
}

// Any returns a Filterer that returns true if any of the
// fs return true.
func Any(fs ...Filterer) Filterer {
  ors := make([][]Filterer, len(fs))
  for i := range fs {
    ors[i] = orList(fs[i])
  }
  return orFilterer(filterFlatten(ors))
}

// All returns a Filterer that returns true if all of the
// fs return true.
func All(fs ...Filterer) Filterer {
  ands := make([][]Filterer, len(fs))
  for i := range fs {
    ands[i] = andList(fs[i])
  }
  return andFilterer(filterFlatten(ands))
}

// Compose composes two Mappers together into one e.g f(g(x)). c
// creates instances to hold the results of g. Returned Mapper is
// thread-safe.
func Compose(f Mapper, g Mapper, c Creater) Mapper {
  l := mapperLen(f) + mapperLen(g)
  mappers := make([]Mapper, l)
  creaters := make([]Creater, l - 1)
  n := appendMapper(mappers, creaters, g)
  creaters[n - 1] = c
  appendMapper(mappers[n:], creaters[n:], f)
  return &compositeMapper{mappers, creaters, nil}
}

// NewFilterer returns a new Filter
func NewFilterer(f func(ptr interface{}) bool) Filterer {
  return funcFilterer(f)
}

// NewMapper returns a new Mapper
func NewMapper(m func(srcPtr interface{}, destPtr interface{}) bool) Mapper {
  return funcMapper(m)
}

// NewCreater returns a creater that simply returns a pointer to new storage.
// returned pointer is of same type as ptr.
func NewCreater(ptr interface{}) Creater {
  valueType := reflect.TypeOf(ptr).Elem()
  return simpleCreater{valueType}
}

// NewCreater from func returns a creater that delegates to f.
func NewCreaterFromFunc(f func() interface{}) Creater {
  return funcCreater{f}
}

type count struct {
  start int
  step int
}

func (c *count) Next(ptr interface{}) bool {
  p := ptr.(*int)
  *p = c.start
  c.start += c.step
  return true
}

type mapStream struct {
  mapper Mapper
  stream Stream
  ptr interface{} 
}

func (s *mapStream) Next(ptr interface{}) bool {
  for s.stream.Next(s.ptr) {
    if s.mapper.Map(s.ptr, ptr) {
      return true
    }
  }
  return false
}

type filterStream struct {
  filterer Filterer
  stream Stream
}

func (s *filterStream) Next(ptr interface{}) bool {
  for s.stream.Next(ptr) {
    if s.filterer.Filter(ptr) {
      return true
    }
  }
  return false
}

type sliceStream struct {
  stream Stream
  start int
  end int
  index int
}

func (s *sliceStream) Next(ptr interface{}) bool {
  for (s.end < 0 || s.index < s.end) && s.stream.Next(ptr) {
    if s.index >= s.start {
      s.index++
      return true
    }
    s.index++
  }
  return false
}

type concatStream struct {
  streams []Stream
  index int
}

func (s *concatStream) Next(ptr interface{}) bool {
  for s.index < len(s.streams) && !s.streams[s.index].Next(ptr) {
    s.index++
  }
  return s.index < len(s.streams)
}

type joinStream struct {
  streams []Stream
  done bool
}

func (s *joinStream) Next(ptr interface{}) bool {
  if s.done {
    return false
  }
  ptrs := ptr.(Tuple).Ptrs()
  for i := range s.streams {
    if !s.streams[i].Next(ptrs[i]) {
      s.done = true
      return false
    }
  }
  return true
}

type cycleStream struct {
  sliceValue reflect.Value
  length int
  index int
}

func (s *cycleStream) Next(ptr interface{}) bool {
  value := s.sliceValue.Index(s.index % s.length)
  reflect.Indirect(reflect.ValueOf(ptr)).Set(value)
  s.index++
  return true
}

type takeStream struct {
  filterer Filterer
  stream Stream
  done bool
}

func (s *takeStream) Next(ptr interface{}) bool {
  for !s.done && s.stream.Next(ptr) {
    if s.filterer.Filter(ptr) {
      return true
    }
    s.done = true
  }
  return false
}

type dropStream struct {
  filterer Filterer
  stream Stream
  done bool
}

func (s *dropStream) Next(ptr interface{}) bool {
  for s.stream.Next(ptr) {
    if s.done {
      return true
    }
    if !s.filterer.Filter(ptr) {
      s.done = true
      return true
    }
  }
  return false
}

type lineStream struct {
  *bufio.Reader
}

func (s lineStream) Next(ptr interface{}) bool {
  p := ptr.(*string)
  line, isPrefix, err := s.ReadLine()
  if err == io.EOF {
    return false
  }
  if err != nil {
    panic(fmt.Sprintf("Received unexpected error %v", err))
  }
  if !isPrefix {
    *p = string(line)
    return true
  }
  *p = s.readRestOfLine(line)
  return true
}

func (s lineStream) readRestOfLine(line []byte) string {
  lines := [][]byte{copyBytes(line)}
  for {
    l, isPrefix, err := s.ReadLine()
    if err == io.EOF {
      break
    }
    if err != nil {
      panic(fmt.Sprintf("Received unexpected error %v", err))
    }
    lines = append(lines, copyBytes(l))
    if !isPrefix {
      break
    }
  }
  return string(byteFlatten(lines))
}

type rowStream struct {
  Rows
}

func (r rowStream) Next(ptr interface{}) bool {
  if !r.Rows.Next() {
    return false
  }
  ptrs := ptr.(Tuple).Ptrs()
  if err := r.Scan(ptrs...); err != nil {
    panic(err)
  }
  return true
}

type funcFilterer func(ptr interface{}) bool

func (f funcFilterer) Filter(ptr interface{}) bool {
  return f(ptr)
}

type andFilterer []Filterer

func (f andFilterer) Filter(ptr interface{}) bool {
  for i := range f {
    if !f[i].Filter(ptr) {
      return false
    }
  }
  return true
}

type orFilterer []Filterer

func (f orFilterer) Filter(ptr interface{}) bool {
  for i := range f {
    if f[i].Filter(ptr) {
      return true
    }
  }
  return false
}

type funcMapper func(srcPtr interface{}, destPtr interface{}) bool

func (m funcMapper) Map(srcPtr interface{}, destPtr interface{}) bool {
  return m(srcPtr, destPtr)
}

func (m funcMapper) Fast() Mapper {
  return m
}

type compositeMapper struct {
  mappers []Mapper
  creaters []Creater
  values []interface{}
}

func (m *compositeMapper) Map(srcPtr interface{}, destPtr interface{}) bool {
  if m.values != nil {
    num := len(m.mappers)
    if !m.mappers[0].Map(srcPtr, m.values[0]) {
      return false
    }
    for i := 1; i < num - 1; i++ {
      if !m.mappers[i].Map(m.values[i-1], m.values[i]) {
        return false
      }
    }
    if !m.mappers[num - 1].Map(m.values[num - 2], destPtr) {
      return false
    }
    return true
  }
  return m.Fast().Map(srcPtr, destPtr)
}

func (m *compositeMapper) Fast() Mapper {
  if m.values != nil {
    return m
  }
  return &compositeMapper{m.mappers, m.creaters, m.createValues()}
}

func (m *compositeMapper) createValues() []interface{} {
  result := make([]interface{}, len(m.creaters))
  for i := range m.creaters {
    result[i] = m.creaters[i].Create()
  }
  return result
}

type simpleCreater struct {
  reflect.Type
}

func (c simpleCreater) Create() interface{} {
  return reflect.New(c.Type).Interface()
}

type funcCreater struct {
  f func() interface{}
}

func (c funcCreater) Create() interface{} {
  return c.f()
}

func appendPtrs(s Stream, c Creater, sliceValue reflect.Value) reflect.Value {
  value := c.Create()
  for s.Next(value) {
    sliceValue = reflect.Append(sliceValue, reflect.ValueOf(value))
    value = c.Create()
  }
  return sliceValue
}

func appendValues(s Stream, sliceElementType reflect.Type, sliceValue reflect.Value) reflect.Value {
  value := reflect.New(sliceElementType)
  for s.Next(value.Interface()) {
    sliceValue = reflect.Append(sliceValue, reflect.Indirect(value))
  }
  return sliceValue
}

func getSliceValue(slicePtr interface{}) reflect.Value {
  sliceValue := reflect.Indirect(reflect.ValueOf(slicePtr))
  if !sliceValue.CanSet() || sliceValue.Kind() != reflect.Slice {
    panic("slicePtr must be a pointer to a slice.")
  }
  return sliceValue
}

func orList(f Filterer) []Filterer {
  ors, ok := f.(orFilterer)
  if ok {
    return ors
  }
  return []Filterer{f}
}

func andList(f Filterer) []Filterer {
  ands, ok := f.(andFilterer)
  if ok {
    return ands
  }
  return []Filterer{f}
}

func filterFlatten(fs [][]Filterer) []Filterer {
  var l int
  for i := range fs {
    l += len(fs[i])
  }
  result := make([]Filterer, l)
  n := 0
  for i := range fs {
    n += copy(result[n:], fs[i])
  }
  return result
}

func mapperLen(m Mapper) int {
  cm, ok := m.(*compositeMapper)
  if ok {
    return len(cm.mappers)
  }
  return 1
}

func appendMapper(mappers []Mapper, creaters []Creater, m Mapper) int {
  cm, ok := m.(*compositeMapper)
  if ok {
    copy(creaters, cm.creaters)
    return copy(mappers, cm.mappers)
  }
  mappers[0] = m
  return 1
}

func copyBytes(b []byte) []byte {
  result := make([]byte, len(b))
  copy(result, b)
  return result
}

func byteFlatten(b [][]byte) []byte {
  var l int
  for i := range b {
    l += len(b[i])
  }
  result := make([]byte, l)
  n := 0
  for i := range b {
    n += copy(result[n:], b[i])
  }
  return result
}

