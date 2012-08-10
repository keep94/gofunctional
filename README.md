# gofunctional

functional programming in go. The main data type, Stream, is similar to
a python iterator or generator. The methods found in here are similar to
the methods found in the python itertools module.

	import "github.com/keep94/gofunctional/functional"

## Installing

	go get github.com/keep94/gofunctional

## Real World Example

Suppose there are names and phone numbers of people stored in a sqlite
database. The table has a name, and phone_number column.

The person class would look like:

	type Person struct {
	  Name string
	  Phone string
	}

	func (p *Person) Ptrs() {
	  return []interface{}{&p.Name, &p.Phone}
	}

To get the 4th page of 25 people do:

	package main

	import (
	  "code.google.com/p/gosqlite/sqlite"
	  "github.com/keep94/gofunctional/functional"
	)

	func main() {
	  conn, _ := sqlite.Open("YourDataFilePath")
	  stmt, _ := conn.Prepare("select * from People")
	  s := functional.ReadRows(stmt)
	  s = functional.Slice(s, 3 * 25, 4 * 25)
	  for person := new(Person); s.Next(person); person = new(Person) {
	    // Display person here
	  }
	}

To store the 4th page of people in a slice, the for loop above can be
replaced with:

	people []*Person
	functional.AppendPtrs(s, &people, nil)
	// Do something with people slice

Like python iterators and generators, Stream types are lazily evaluated, so
the above code will read only the first 100 names no matter how many people
are in the database.

See tests and the included example for detailed usage.

