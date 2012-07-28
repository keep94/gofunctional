package main

// This program demonstrates using the functional package to perform operations
// on the fibonacci sequence. In particular this program computes the ratio
// between the size of fibonacci numbers and their position in the sequence.

import (
  "fmt"
  "github.com/keep94/gofunctional/functional"
  "math/big"
)

// fibonacci returns the fibonacci numbers as a stream of big.Int. 
func fibonacci() functional.Stream {
  return &fibStream{big.NewInt(0), big.NewInt(1)}
}

type fibStream struct {
  a *big.Int
  b *big.Int
}

func (s *fibStream) Next(ptr interface{}) bool {
  p := ptr.(*big.Int)
  p.Set(s.a)
  s.a.Set(s.b)
  s.b.Add(p, s.a)
  return true
}

// fibWithIndex contains the position of the fibonacci number
// along with the fibonacci number. Note that this implements
// functional.Tuple so that it can receive values that functional.Join
// emits
type fibWithIndex struct {
  idx int
  num *big.Int
}

func (f *fibWithIndex) Ptrs() []interface{} {
  return []interface{}{&f.idx, f.num}
}

// ratioWithIndex contains the position of the fibonacci number along
// with the ratio of its size to its position
type ratioWithIndex struct {
  idx int
  ratio float64
}

// computeRatio maps a fibWithIndex instance to a ratioWithIndex
// instance.
func computeRatio(srcPtr, destPtr interface{}) bool {
  p := srcPtr.(*fibWithIndex)
  q := destPtr.(*ratioWithIndex)
  q.idx = p.idx
  q.ratio = float64(p.num.BitLen()) / float64(p.idx)
  return true
}

func main() {
  s := fibonacci()
  s = functional.Join(functional.Count(), s)
  s = functional.Map(
          functional.NewMapper(computeRatio),
          s, 
          functional.NewCreaterFromFunc(func() interface{} {
           return &fibWithIndex{0, new(big.Int)}
          }))
  
  // Index and ratio from 40th up to 49th fibonacci number.
  s = functional.Slice(s, 40, 50)
  var results []ratioWithIndex
  functional.AppendValues(s, &results)
  fmt.Println(results)
}
  
