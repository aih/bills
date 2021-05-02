package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/aih/bills"
)

/*
func main() {
	bills.CompareSamples()
}
*/

// BillList is a string slice
type BillList []string

func (bl *BillList) String() string {
	return fmt.Sprintln(*bl)
}

// Set string value in MyList
func (bl *BillList) Set(s string) error {
	// TODO remove leading and following spaces and
	// split on other spaces
	*bl = strings.Split(s, ",")
	return nil
}

func main() {
	flagPathUsage := "Absolute path to the parent directory for 'congress' and json metadata files"
	flagPathValue := string(bills.ParentPathDefault)
	var parentPath string
	flag.StringVar(&parentPath, "parentPath", flagPathValue, flagPathUsage)
	flag.StringVar(&parentPath, "p", flagPathValue, flagPathUsage+" (shorthand)")

	var billList BillList
	flag.Var(&billList, "billnumbers", "comma-separated list of billnumbers")
	flag.Parse()

	bills.CompareBills(parentPath, billList)
}
