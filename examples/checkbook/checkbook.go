// checkbook is a small program that prints a checkbook register from a
// database. It first reads from the database the balance of the account.
// Then it prints the entries in the account showing the balance at each
// transaction. This program creates a Generator to emit the
// entries with the current balance.
//
// When this program is run, the current working directory needs to be this
// directory or else the program will not find the sqlite file, chkbook.db
package main

import (
  "code.google.com/p/gosqlite/sqlite"
  "errors"
  "fmt"
  "github.com/keep94/gofunctional/functional"
)

// Entry represents an entry in a checkbook register
type Entry struct {
  // YYYYmmdd format
  Date string
  Name string
  // $40.64 is 4064
  Amount int64
  // Balance is the remaining balance in account. $40.64 is 4064
  Balance int64
}

func (e *Entry) String() string {
  return fmt.Sprintf("date: %s; name: %s; amount: %d; balance: %d", e.Date, e.Name, e.Amount, e.Balance)
}

func (e *Entry) Ptrs() []interface{} {
  return []interface{} {&e.Date, &e.Name, &e.Amount}
}
  
// ChkbookEntries returns a Generator that emits all the entries in a
// checkbook ordered by most recent to least recent. conn is the sqlite
// connection; acctId is the id of the account for which to print entries.
// If acctId does not match a valid account, ChkbookEntries will return an
// error and nil for the Generator. If caller does not exhaust returned
// Generator, it must call Close on it to free up resources.
func ChkbkEntries(conn *sqlite.Conn, acctId int) (functional.Generator, error) {
  stmt, err := conn.Prepare("select balance from balances where acct_id = ?")
  if err != nil {
   return nil, err
  }
  if err = stmt.Exec(acctId); err != nil {
    stmt.Finalize()
    return nil, err
  }
  if !stmt.Next() {
    stmt.Finalize()
    return nil, errors.New("No balance")
  }
  var bal int64
  if err = stmt.Scan(&bal); err != nil {
    stmt.Finalize()
    return nil, err
  }
  stmt.Finalize()
  stmt, err = conn.Prepare("select date, name, amount from entries where acct_id = ? order by date desc")
  if err != nil {
    return nil, err
  }
  if err = stmt.Exec(acctId); err != nil {
    stmt.Finalize()
    return nil, err
  }
  return functional.NewGenerator(func(emitter functional.Emitter) {
    rowStream := functional.ReadRows(stmt)
    for ptr := emitter.EmitPtr(); ptr != nil && rowStream.Next(ptr); ptr = emitter.EmitPtr() {
      entry := ptr.(*Entry)
      entry.Balance = bal
      bal += entry.Amount
    }
    stmt.Finalize()
  }), nil
}

func main() {
  conn, err := sqlite.Open("chkbook.db")
  if err != nil {
    fmt.Println("Error opening file")
    return
  }
  g, err := ChkbkEntries(conn, 1)
  if err != nil {
    fmt.Printf("Error reading ledger %v", err)
  }
  var entry Entry
  for g.Next(&entry) {
    fmt.Println(&entry)
  }
  // Since we exhaust g we don't need to close explicitly, but it is good
  // practice to always close a Generator
  g.Close()
}
