// This program solves the following puzzle: Find the nth digit of the
// sequence 12345678910111213141516... To find the 287th digit, run "puzzle 287"
package main

import (
  "flag"
  "fmt"
  "github.com/keep94/gofunctional/functional"
  "strconv"
)

// Emits 01234567891011121314151617... as a Stream of runes.
func AllDigits() functional.Generator {
  return functional.NewGenerator(
      func(e functional.Emitter) {
        for number := 0; ; number++ {
          for _, ch := range strconv.Itoa(number) {
            ptr := e.EmitPtr()
            if ptr == nil {
              return
            }
            *(ptr.(*rune)) = ch
          }
        }
      })
}

// Return the nth digit of 1234567891011121314151617....
func Digit(posit int) string {
  g := AllDigits()
  s := functional.Slice(g, posit, -1)
  var r rune
  s.Next(&r)
  g.Close()
  return string(r)
}

func main() {
  flag.Parse()
  posit, _ := strconv.Atoi(flag.Arg(0))
  fmt.Println(Digit(posit))
}
