// Package functional provides functional programming constructs.
package functional

import (
  "bufio"
  "fmt"
  "io"
  "reflect"
)

// Stream is a sequence emitted values.
// Each call to Next() emits the next value in the stream.
// A Stream that emits values of type T is a Stream of T.
// Because the client of a Stream of T may need save only a small
// subset of the emitted values, to save the Stream from allocating memory
// for each emitted T value, it is the client's responsibility to 
// allocate memory to store the next emitted value. The client passes a
// pointer, a *T, to Next to receive the next emitted value. If the
// client does not want to store the previously emitted values, it is free 
// to re-use the same *T with Next over and over again. If type T includes
// resources that need to be pre-initialized, a Stream may require the client
// to pass a pointer to a pre-initialized T to save the Stream from having to
// initialize a T for every emitted value. Clients use the Creater type to
// initialize a new T and get a *T pointer to it.
type Stream interface {
  // Next emits the next value in this Stream of T.
  // If Next returns true, the next value is stored at ptr.
  // If Next returns false, then the end of the Stream has been reached,
  // and the value ptr points to is unspecified. ptr must be a *T
  Next(ptr interface{}) bool
}

// Tuple represents a tuple of values that Join emits
type Tuple interface {
  // Ptrs returns a pointer to each field in the tuple.
  Ptrs() []interface{}
}

// Filterer of T filters values in a Stream of T.
type Filterer interface {
  // Filter returns true if value ptr points to should be included or false
  // otherwise. ptr must be a *T.
  Filter(ptr interface{}) bool
}

// Mapper maps a type T value to a type U value in a Stream.
type Mapper interface {
  // Map does the mapping storing the mapped value at destPtr.
  // If Mapper returns false, then no mapped value is stored at destPtr.
  // srcPtr is a *T; destPtr is a *U
  Map(srcPtr interface{}, destPtr interface{}) bool
  // Fast returns a faster version of this Mapper. If a function will use
  // a Mapper more than once, say in a for loop, it should call Fast and use
  // the returned Mapper instead. Returned Mapper should be considered not
  // thread-safe even if this Mapper is. In particular, the returned Mapper
  // may re-use temporary storage rather than creating it anew each time Map
  // is invoked. Most implementations can simply return themselves.
  Fast() Mapper
}

// Creater of T creates a new, pre-initialized, T and returns a pointer to it.
type Creater func() interface {}

// Copier of T copies the value at src to the value at dest. This type is
// often needed when values of type T need to be pre-initialized. src and
// dest are of type *T and both point to pre-initialized T.
type Copier func(src, dest interface{})

// KeyFunc of T returns a key for a type T. The key type should support
// both equality and assignment e.g int, string. ptr is a type *T. 
type KeyFunc func(ptr interface{}) interface{}

// Rows represents rows in a database table. Most database API already have
// a type that implements this interface
type Rows interface {
  // Next advances to the next row. Next returns false if there is no next row.
  // Every call to Scan, even the first one, must be preceded by a call to Next.
  Next() bool
  // Reads the values out of the current row. args are pointer types.
  Scan(args ...interface{}) error
}

// Group of T is a Stream of T that have a common key.
type Group struct {
  s Stream
  key interface{}
  ptr interface{}
  k KeyFunc
  c Copier
  keySet bool
  ptrSaved bool
  halted bool
}

// Next emits the next value of type T. ptr is a *T.
// If there are no more values, Next returns false.
func (g *Group) Next(ptr interface{}) bool {
  if g.halted {
    return false
  }
  if g.ptrSaved {
    g.copyValue(g.ptr, ptr)
    g.ptrSaved = false
    return true
  }
  if g.s.Next(ptr) {
    if !g.isSameKey(g.k(ptr)) {
      g.copyValue(ptr, g.ptr)
      g.ptrSaved = true
      g.halted = true
      return false
    }
    return true
  }
  return false
}
      
// Key returns the common key for this Group
func (g *Group) Key() interface{} {
  return g.key
}

func (g *Group) copyValue(src, dest interface{}) {
  if src == dest {
    return
  }
  g.c(src, dest)
}

func (g *Group) isSameKey(key interface{}) bool {
  return g.keySet && g.key == key
}

func (g *Group) advance() bool {
  for g.Next(g.ptr) {
  }
  if g.halted {
    g.halted = false
    g.key = g.k(g.ptr)
    g.keySet = true
    return true
  }
  return false
}

// Map applies f, which maps a type T value to a type U value, to a Stream
// of T producing a new Stream of U. If s is
// (x1, x2, x3, ...), Map returns the Stream (f(x1), f(x2), f(x3), ...).
// if f returns false for a T value, then the corresponding U value is left
// out of the returned stream. ptr is a *T providing storage for emitted values
// from s. Clients need not pass f.Fast() to Map because Map calls Fast
// internally.
func Map(f Mapper, s Stream, ptr interface{}) Stream {
  ms, ok := s.(*mapStream)
  if ok {
    return &mapStream{Compose(f, ms.mapper, newCreater(ptr)).Fast(), ms.stream, ms.ptr}
  }
  return &mapStream{f.Fast(), s, ptr}
}

// Filter filters values from s, returning a new Stream of T.
// f is a Filterer of T; s is a Stream of T.
func Filter(f Filterer, s Stream) Stream {
  fs, ok := s.(*filterStream)
  if ok {
    return &filterStream{All(fs.filterer, f), fs.stream}
  }
  return &filterStream{f, s}
}

// Count returns an infinite Stream of int which emits all values beginning
// at 0.
func Count() Stream {
  return &count{0, 1}
}

// CountFrom returns an infinite Stream of int emitting values beginning at
// start and increasing by step.
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
  return Flatten(NewStreamFromValues(s))
}

// Join uses multiple Streams to form a new Stream of Tuples.
// if x = (x1, x2, ..) and y = (y1, y2, ...) then
// Join(x, y) = ((x1, y1), (x2, y2), ...). 
// The Stream Join returns quits emitting whenever one of the input Streams
// runs out.
func Join(s ...Stream) Stream {
  return &joinStream{s}
}

// Cycle is deprecated. See CycleValues
func Cycle(aSlice interface{}) Stream {
  return CycleValues(aSlice)
}

// CycleValues emits the elements in the given slice over and over again.
// CycleValues([]int {3, 5}) = (3, 5, 3, 5, ...). If CycleValues is passed
// an empty slice, then CycleValues returns an empty Stream. If aSlice is a
// []T then CycleValues returns a Stream of T.
func CycleValues(aSlice interface{}) Stream {
  sliceValue := getSliceValueFromValue(aSlice)
  return &cycleStream{sliceValue, assignFromValue, sliceValue.Len(), 0}
}

// CyclePtrs is like CycleValues except aSlice []*T slice, and CyclePtrs
// returns a Stream of T. c is a Copier of T. If c is nil, regular assignment
// is used.
func CyclePtrs(aSlice interface{}, c Copier) Stream {
  sliceValue := getSliceValueFromValue(aSlice)
  assertPtrType(sliceValue.Type().Elem())
  return &cycleStream{sliceValue, toSliceValueCopy(c), sliceValue.Len(), 0}
}

// NewStreamFromValues convers a []T into a Stream of T. aSlice is a []T.
func NewStreamFromValues(aSlice interface{}) Stream {
  sliceValue := getSliceValueFromValue(aSlice)
  return &plainStream{sliceValue, assignFromValue, sliceValue.Len(), 0}
}

// NewStreamFromPtrs converts a []*T into a Stream of T. aSlice is a []*T.
// c is a Copier of T. if c is nil, regular assignment is used.
func NewStreamFromPtrs(aSlice interface{}, c Copier) Stream {
  sliceValue := getSliceValueFromValue(aSlice)
  assertPtrType(sliceValue.Type().Elem())
  return &plainStream{sliceValue, toSliceValueCopy(c), sliceValue.Len(), 0}
}

// NilStream returns a stream that emits no values.
func NilStream() Stream {
  return kNilStream
}

// Flatten converts a Stream of Stream of T into a Stream of T.
func Flatten(s Stream) Stream {
  return &flattenStream{stream: s}
}

// TakeWhile returns a Stream that emits the values in s until f is false.
// f is a Filterer of T; s is a Stream of T.
func TakeWhile(f Filterer, s Stream) Stream {
  return &takeStream{f, s}
}

// DropWhile returns a Stream that emits the values in s starting at the
// first value where f is false. f is a Filterer of T; s is a Stream of T.
func DropWhile(f Filterer, s Stream) Stream {
  return &dropStream{f, s}
}

// ReadLines returns the lines of text in r separated by either "\n" or "\r\n"
// as a Stream of string. The emitted string types do not contain the
// end of line characters.
func ReadLines(r io.Reader) Stream {
  return lineStream{bufio.NewReader(r)}
}

// ReadRows returns the rows in a database table as a Stream of Tuple.
func ReadRows(r Rows) Stream {
  return rowStream{r}
}

// PartitionValues converts a Stream of T to a Stream of []T where each
// emitted value has same length. When calling Next on the returned Stream,
// pass a pointer pointing to a slice of type []T initialized with make.
// The returned Stream fills this slice with values with each call to Next.
// If s runs out before returned Stream can completly fill the slice, it
// emits a smaller slice with just the remaining values to the pointer
// passed to Next.
func PartitionValues(s Stream) Stream {
  return partitionValuesStream{s}
}

// PartitionPtrs converts a Stream of T to a Stream of []*T where each
// emitted value has same length. When calling Next on returned Stream,
// pass a pointer pointing to a slice of type []*T initialized with make
// and InitPtrs. The returned Stream fills this slice with values with
// each call to Next. If s runs out before returned Stream can completely
// fill the slice, it emits a smaller slice with just the remaining values
// to the pointer passed to Next.
func PartitionPtrs(s Stream) Stream {
  return partitionPtrsStream{s}
}

// GroupBy returns a Stream of *Group that emits the T values in s grouped
// by key. k is applied to each element in s to determine its key.
// Each *Group instance is itself a Stream of T that emits all the values with
// a given key. Each *Group instance is valid until Next is called again
// on the returned Stream. Threfore callers should discard any previously
// emitted *Group values. The values in s must already be sorted by k.
// s must not be used directly once this function is called. k is the
// key generating funtion which takes a *T pointer and returns the key.
// ptr is a *T pointer providing storage for emitted values from s.
// c is a Copier of T. If c is nil, it means use the assignment operator.
func GroupBy(s Stream, k KeyFunc, ptr interface{}, c Copier) Stream {
  if c == nil {
    c = assignCopier
  }
  return groupByStream{&Group{s: s, ptr: ptr, k: k, c: c}}
}

// Deferred returns a Stream that emits the values from the Stream f returns.
// f is not called until the first time Next is called on the returned stream.
// In this way, the creation of a Stream can be deferred until the values
// it emits are needed.
func Deferred(f func() Stream) Stream {
  return &deferredStream{f, nil}
}

// AppendValues evaluates s and appends each element in s to the slice that
// slicePtr points to. s is a Stream of T; slicePtr is a *[]T
func AppendValues(s Stream, slicePtr interface{}) {
  sliceValue := getSliceValueFromPtr(slicePtr)
  sliceElementType := sliceValue.Type().Elem()
  sliceValue.Set(appendValues(s, sliceElementType, sliceValue))
}

// AppendPtrs evaluates s and appends each element in s to the slice that
// slicePtr points to. s is a Stream of T; slicePtr is a *[]*T.
// c is a Creater of T. If c is nil, it means use the new built-in function.
func AppendPtrs(s Stream, slicePtr interface{}, c Creater) {
  sliceValue := getSliceValueFromPtr(slicePtr)
  sliceElementType := sliceValue.Type().Elem()
  assertPtrType(sliceElementType)
  if c == nil {
    sliceValue.Set(appendPtrs(s, sliceElementType.Elem(), sliceValue))
  } else {
    sliceValue.Set(appendPtrsWithCreater(s, c, sliceValue))
  }
}

// CopyValues copies emitted values from s to aSlice until either s
// is exhausted or until it reaches the end of aSlice. CopyValues
// returns the number of emitted values copied. If end of aSlice isn't
// reached, caller must not assume anything about the contents of the rest of
// the slice. s is a Stream of T; aSlice is a []T.
func CopyValues(s Stream, aSlice interface{}) int {
  sliceValue := getSliceValueFromValue(aSlice)
  return copyToSlice(s, sliceValue, valueToInterface)
}

// CopyPtrs copies emitted values from s to aSlice until either s
// is exhausted or until it reaches the end of aSlice. CopyPtrs
// returns the number of emitted values copied. If end of aSlice isn't
// reached, caller must not assume anything about the contents of the rest of
// the slice. s is a Stream of T; aSlice is a []*T.
// aSlice must be pre-initialized with InitPtrs.
func CopyPtrs(s Stream, aSlice interface{}) int {
  sliceValue := getSliceValueFromValue(aSlice)
  assertPtrType(sliceValue.Type().Elem())
  return copyToSlice(s, sliceValue, ptrToInterface)
}

// InitPtrs initializes aSlice of type []*T
// by having each element point to a new T.  c is a 
// Creater of T. If c is nil, new(T) is used to create each T instance.
// InitPtrs returns the same slice passed to it.
func InitPtrs(aSlice interface{}, c Creater) interface{} {
  sliceValue := getSliceValueFromValue(aSlice)
  sliceElementType := sliceValue.Type().Elem()
  assertPtrType(sliceElementType)
  length := sliceValue.Len()
  var creater func() reflect.Value
  if c != nil {
    creater = func() reflect.Value { return reflect.ValueOf(c()) }
  } else {
    sliceElementValueType := sliceElementType.Elem()
    creater = func() reflect.Value { return reflect.New(sliceElementValueType) }
  }
  for i := 0; i < length; i++ {
    sliceValue.Index(i).Set(creater())
  }
  return aSlice
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

// Compose composes two Mappers together into one e.g f(g(x)). If g maps
// type T values to type U values, and f maps type U values to type V
// values, then Compose returns a Mapper mapping T values to V values. c is
// a Creater of U. Each time Map is called on returned Mapper, it invokes c
// to create a U value to receive the intermediate result from g. Calling
// Fast() on returned Mapper creates a new Mapper with this U value already
// pre-initialized.
func Compose(f Mapper, g Mapper, c Creater) Mapper {
  l := mapperLen(f) + mapperLen(g)
  mappers := make([]Mapper, l)
  creaters := make([]Creater, l - 1)
  n := appendMapper(mappers, creaters, g)
  creaters[n - 1] = c
  appendMapper(mappers[n:], creaters[n:], f)
  return &compositeMapper{mappers, creaters, nil}
}

// NewFilterer returns a new Filterer of T. f takes a *T returning true
// if T value pointed to it should be included.
func NewFilterer(f func(ptr interface{}) bool) Filterer {
  return funcFilterer(f)
}

// NewMapper returns a new Mapper mapping T values to U Values. In f,
// srcPtr is a *T and destPtr is a *U pointing to pre-allocated T and U
// values respectively.
func NewMapper(m func(srcPtr interface{}, destPtr interface{}) bool) Mapper {
  return funcMapper(m)
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

type flattenStream struct {
  stream Stream
  current Stream
}

func (s *flattenStream) Next(ptr interface{}) bool {
  if s.stream == nil {
    return false
  }
  for s.current == nil || !s.current.Next(ptr) {
    if !s.stream.Next(&s.current) {
      s.stream = nil
      return false
    }
  }
  return true
}

type joinStream struct {
  streams []Stream
}

func (s *joinStream) Next(ptr interface{}) bool {
  if s.streams == nil {
    return false
  }
  ptrs := ptr.(Tuple).Ptrs()
  for i := range s.streams {
    if !s.streams[i].Next(ptrs[i]) {
      s.streams = nil
      return false
    }
  }
  return true
}

type cycleStream struct {
  sliceValue reflect.Value
  copyFunc func(src reflect.Value, dest interface{})
  length int
  index int
}

func (s *cycleStream) Next(ptr interface{}) bool {
  if s.length == 0 {
    return false
  }
  s.copyFunc(s.sliceValue.Index(s.index), ptr)
  s.index = (s.index + 1) % s.length
  return true
}

type plainStream struct {
  sliceValue reflect.Value
  copyFunc func(src reflect.Value, dest interface{})
  length int
  index int
}

func (s *plainStream) Next(ptr interface{}) bool {
  if s.index == s.length {
    return false
  }
  s.copyFunc(s.sliceValue.Index(s.index), ptr)
  s.index++
  return true
}

type takeStream struct {
  filterer Filterer
  stream Stream
}

func (s *takeStream) Next(ptr interface{}) bool {
  for s.stream != nil && s.stream.Next(ptr) {
    if s.filterer.Filter(ptr) {
      return true
    }
    s.stream = nil
  }
  return false
}

type dropStream struct {
  filterer Filterer
  stream Stream
}

func (s *dropStream) Next(ptr interface{}) bool {
  for s.stream.Next(ptr) {
    if s.filterer == nil {
      return true
    }
    if !s.filterer.Filter(ptr) {
      s.filterer = nil
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

type partitionValuesStream struct {
  Stream
}

func (s partitionValuesStream) Next(slicePtr interface{}) bool {
  sliceValue := getSliceValueFromPtr(slicePtr)
  return nextSlice(s.Stream, sliceValue, valueToInterface)
}

type partitionPtrsStream struct {
  Stream
}

func (s partitionPtrsStream) Next(slicePtr interface{}) bool {
  sliceValue := getSliceValueFromPtr(slicePtr)
  assertPtrType(sliceValue.Type().Elem())
  return nextSlice(s.Stream, sliceValue, ptrToInterface)
}

type groupByStream struct {
  *Group
}

func (g groupByStream) Next(ptr interface{}) bool {
  if !g.advance() {
    return false
  }
  p := ptr.(**Group)
  *p = g.Group
  return true
}

type deferredStream struct {
  f func() Stream
  s Stream
}

func (d *deferredStream) Next(ptr interface{}) bool {
  if d.s == nil {
    d.s = d.f()
  }
  return d.s.Next(ptr)
}

type nilStream struct {
}

func (s nilStream) Next(ptr interface{}) bool {
  return false
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
  return &compositeMapper{m.fastMappers(), m.creaters, m.createValues()}
}

func (m *compositeMapper) createValues() []interface{} {
  result := make([]interface{}, len(m.creaters))
  for i := range m.creaters {
    result[i] = m.creaters[i]()
  }
  return result
}

func (m *compositeMapper) fastMappers() []Mapper {
  result := make([]Mapper, len(m.mappers))
  for i := range m.mappers {
    result[i] = m.mappers[i].Fast()
  }
  return result
}

func appendPtrsWithCreater(
    s Stream, c Creater, sliceValue reflect.Value) reflect.Value {
  value := c()
  for s.Next(value) {
    sliceValue = reflect.Append(sliceValue, reflect.ValueOf(value))
    value = c()
  }
  return sliceValue
}

func appendPtrs(
    s Stream,
    sliceElementType reflect.Type,
    sliceValue reflect.Value) reflect.Value {
  value := reflect.New(sliceElementType)
  for s.Next(value.Interface()) {
    sliceValue = reflect.Append(sliceValue, value)
    value = reflect.New(sliceElementType)
  }
  return sliceValue
}

func appendValues(
    s Stream,
    sliceElementType reflect.Type,
    sliceValue reflect.Value) reflect.Value {
  value := reflect.New(sliceElementType)
  for s.Next(value.Interface()) {
    sliceValue = reflect.Append(sliceValue, reflect.Indirect(value))
  }
  return sliceValue
}

func getSliceValueFromPtr(slicePtr interface{}) reflect.Value {
  sliceValue := reflect.Indirect(reflect.ValueOf(slicePtr))
  if !sliceValue.CanSet() || sliceValue.Kind() != reflect.Slice {
    panic("slicePtr must be a pointer to a slice.")
  }
  return sliceValue
}

func getSliceValueFromValue(aSlice interface{}) reflect.Value {
  sliceValue := reflect.ValueOf(aSlice)
  if sliceValue.Kind() != reflect.Slice {
    panic("Slice argument expected")
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

func newCreater(ptr interface{}) Creater {
  return func() interface{} {
    return ptr
  }
}

func assertPtrType(t reflect.Type) {
  if t.Kind() != reflect.Ptr {
    panic("slicePtr must point to a slice of pointers.")
  }
}

func nextSlice(s Stream, sliceValue reflect.Value, toInterface func(reflect.Value) interface{}) bool {
  length := sliceValue.Len()
  numCopied := copyToSlice(s, sliceValue, toInterface)
  if numCopied == 0 {
    return false
  }
  if numCopied < length {
    sliceValue.Set(sliceValue.Slice(0, numCopied))
  }
  return true
}

func copyToSlice(s Stream, sliceValue reflect.Value, toInterface func(reflect.Value) interface{}) int {
  length := sliceValue.Len()
  var idx int
  for idx = 0; idx < length; idx++ {
    if !s.Next(toInterface(sliceValue.Index(idx))) {
      break
    }
  }
  return idx
}

func assignCopier(src, dest interface{}) {
  srcP := reflect.ValueOf(src)
  assignFromPtr(srcP, dest)
}

func assignFromValue(srcV reflect.Value, dest interface{}) {
  destP := reflect.ValueOf(dest)
  reflect.Indirect(destP).Set(srcV)
}

func assignFromPtr(srcP reflect.Value, dest interface{}) {
  destP := reflect.ValueOf(dest)
  reflect.Indirect(destP).Set(reflect.Indirect(srcP))
}

func toSliceValueCopy(c Copier) func(src reflect.Value, dest interface{}) {
  if c == nil {
    return assignFromPtr
  }
  return func(src reflect.Value, dest interface{}) {
    c(src.Interface(), dest)
  }
}

func valueToInterface(v reflect.Value) interface{} {
  return v.Addr().Interface()
}

func ptrToInterface(v reflect.Value) interface{} {
  return v.Interface()
}

var kNilStream Stream = nilStream{}
