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
  // If the client has closed the associated Generator, EmitPtr will return
  // nil. When that happens, the function should simply return.
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

type regularGenerator struct {
  ptrCh chan interface{}
  doneCh chan bool
}

func (g *regularGenerator) Next(ptr interface{}) bool {
  if g.ptrCh == nil {
    return false;
  }
  g.ptrCh <- ptr
  g.cleanupIfDone()
  return true
}

func (g *regularGenerator) Close() error {
  g.Next(nil)
  return nil
}

func (g *regularGenerator) EmitPtr() interface{} {
  g.doneCh <- false
  return <-g.ptrCh
}

func (g *regularGenerator) cleanupIfDone() {
  if <-g.doneCh {
    close(g.ptrCh)
    close(g.doneCh)
    g.ptrCh = nil
    g.doneCh = nil
  }
}

func genFuncWrapper(f func(e Emitter), g *regularGenerator) {
  f(g)
  g.doneCh <- true
}
