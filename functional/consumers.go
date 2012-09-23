package functional

// A Consumer of T consumes the T values from a Stream of T.
type Consumer interface {

  // Consume consumes values from Stream s
  Consume(s Stream)
}

// ModifyConsumerStream returns a new Consumer that applies f to its Stream
// and then gives the result to c. If c is a Consumer of T and f takes a
// Stream of U and returns a Stream of T, then ModifyConsumerStream returns a
// Consumer of U.
func ModifyConsumerStream(c Consumer, f func(s Stream) Stream) Consumer {
  return &modifiedConsumerStream{c, f}
}

// MultiConsume consumes the values of s, a Stream of T, sending those T
// values to each Consumer in consumers. MultiConsume consumes values from s
// until either s is exhausted or until no Consumer in consumers is accepting
// values. ptr is a *T that receives the values from s. copier is a Copier
// of T used to copy T values to the Streams sent to each Consumer in
// consumers. Passing null for copier means use simple assignment.
func MultiConsume(s Stream, ptr interface{}, copier Copier, consumers ...Consumer) {
  if copier == nil {
    copier = assignCopier
  }
  streams := make([]*splitStream, len(consumers))
  stillConsuming := false
  for i := range streams {
    streams[i] = &splitStream{ptrCh: make(chan interface{}), nextReturnCh: make(chan bool)}
    go consumerWrapper(streams[i], consumers[i])
    if streams[i].cleanupIfDone() {
      stillConsuming = true
    }
  }
  for stillConsuming && s.Next(ptr) {
    stillConsuming = false
    for i := range streams {
      p := streams[i].currentPtr()
      if p != nil {
        copier(ptr, p)
      }
      if streams[i].nextReturn(true) {
        stillConsuming = true
      }
    }
  }
  for stillConsuming {
    stillConsuming = false
    for i := range streams {
      if streams[i].nextReturn(false) {
        stillConsuming = true
      }
    }
  }
}

type modifiedConsumerStream struct {
  c Consumer
  f func(s Stream) Stream
}

func (mc *modifiedConsumerStream) Consume(s Stream) {
  mc.c.Consume(mc.f(s))
}

func consumerWrapper(s *splitStream, c Consumer) {
  c.Consume(s)
  s.ptrCh <- nil
}

type splitStream struct {
  ptrCh chan interface{}
  nextReturnCh chan bool
  ptr interface{}
}

func (s *splitStream) Next(ptr interface{}) bool {
  s.ptrCh <- ptr
  return <-s.nextReturnCh
}

func (s *splitStream) currentPtr() interface{} {
  return s.ptr
}

func (s *splitStream) nextReturn(returnValue bool) bool {
  if s.nextReturnCh == nil {
    return false
  }
  s.nextReturnCh <- returnValue
  return s.cleanupIfDone()
}

func (s *splitStream) cleanupIfDone() bool {
  s.ptr = <-s.ptrCh
  if s.ptr == nil {
    close(s.ptrCh)
    close(s.nextReturnCh)
    s.ptrCh = nil
    s.nextReturnCh = nil
    return false
  }
  return true
}

