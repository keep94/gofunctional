// Demonstrates a Stream that emits the ugly numbers in order which are
// of the form 2^k * 3^l * 5^m for k,l,m >= 0.
// This program prints all the ugly numbers < 100.
package main

import (
    "fmt"
    "github.com/keep94/gofunctional/functional"
    "github.com/keep94/gofunctional/odqueue"
)

// Ugly returns a Stream of int that emits the ugly numbers in order.
// Pass to this function int[]{2, 3, 5}.
func Ugly(primes []int) functional.Stream {
  q := odqueue.NewQueue()
  elements := make([]*odqueue.Element, len(primes))
  current := q.End()
  for i := range elements {
    elements[i] = current
  }
  // 1 is the first ugly number
  q.Add(1)
  products := make([]int, len(primes))
  copy(products, primes)
  primesCopy := make([]int, len(primes))
  copy(primesCopy, primes)
  return &ugly{q, elements, primesCopy, products, current}
}

type ugly struct {
  // q Contains previous ugly numbers
  q *odqueue.Queue
  // elments[i].Value * primes[i] >= next_number_to_be_emitted
  elements []*odqueue.Element
  // Original primes e.g 2, 3, 5
  primes []int
  // elements[i].Value * primes[i] = products[i]
  products []int
  // current is the next value to be emitted
  current *odqueue.Element
}

func (u *ugly) Next(ptr interface{}) bool {
  p := ptr.(*int)
  if u.current.IsEnd() {
    min := u.minProduct()
    u.q.Add(min)
    u.advance(min)
  }
  *p = u.current.Value.(int)
  u.current = u.current.Next()
  return true
}

// minProduct computes the smallest u.products[i]
func (u *ugly) minProduct() int {
  result := u.products[0] 
  for i := 1; i < len(u.products); i++ {
    if result > u.products[i] {
      result = u.products[i]
    }
  }
  return result
}

// advance advances the u.elements[i] so that
// u.primes[i] * u.elements[i].Value > min for all i.
func (u *ugly) advance(min int) {
  for i := range u.products {
    if u.products[i] == min {
      u.elements[i] = u.elements[i].Next()
      u.products[i] = u.primes[i] * u.elements[i].Value.(int)
    }
  }
}

func main() {
  s := Ugly([]int {2, 3, 5})
  s = functional.TakeWhile(
      functional.NewFilterer(func (ptr interface{}) bool {
        p := ptr.(*int)
        return *p < 100
      }), s)
  var n int
  for s.Next(&n) {
    fmt.Println(n)
  }
}
