package functional

import (
  "io"
)

// Generator is a Stream that can be closed.
type Generator interface {
  Stream
  io.Closer
}

// Emitter allows a function to emit values to an associated Generator.
type Emitter interface {

  // To emit a value, a function calls EmitPtr, converts the returned value
  // to the appropriate pointer type, and writes the value to emit there.
  // The the final call to EmitPtr does not emit a value giving the function
  // opportunity to cancel emitting after calling EmitPtr. Calling
  // EmitPtr again causes the value stored at the previous result of EmitPtr
  // to be emitted. If the client has closed the associated Generator,
  // EmitPtr will return nil. When that happens, the function should simply
  // return.
  EmitPtr() interface{}
}

// NewGenerator creates a new Generator that emits the values from emitting
// function f. When f is through emitting values, it should just return. If
// f gets nil when calling EmitPtr on e it should return immediately as this
// means the Generator was closed.
func NewGenerator(f func(e Emitter)) Generator {
  g := &regularGenerator{make(chan interface{}), make(chan bool)}
  go genFuncWrapper(f, g)
  g.cleanupIfDone()
  return g
}

// StreamToGenerator converts a Stream to a Generator. Closing the returned
// Generator closes c.
func StreamToGenerator(s Stream, c io.Closer) Generator {
  return &simpleGenerator{s, c}
}

type regularGenerator struct {
  ptrCh chan interface{}
  doneCh chan bool
}

func (g *regularGenerator) Next(ptr interface{}) bool {
  if g.ptrCh == nil {
    return false
  }
  g.ptrCh <- ptr
  return g.cleanupIfDone()
}

func (g *regularGenerator) Close() error {
  g.Next(nil)
  return nil
}

func (g *regularGenerator) EmitPtr() interface{} {
  g.doneCh <- false
  return <-g.ptrCh
}

func (g *regularGenerator) cleanupIfDone() bool {
  if <-g.doneCh {
    close(g.ptrCh)
    close(g.doneCh)
    g.ptrCh = nil
    g.doneCh = nil
    return false
  }
  return true
}

type simpleGenerator struct {
  Stream
  io.Closer
}

func genFuncWrapper(f func(e Emitter), g *regularGenerator) {
  f(g)
  g.doneCh <- true
}
