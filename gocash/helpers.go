package gocash

import (
	"fmt"
)

// MaxInt returns max int found any number of arguments
func MaxInt(args ...int) (max int) {
	for _, v := range args {
		if v > max {
			max = v
		}
	}
	return
}

// CheckArgs validates `expected` number of args in `got`
func CheckArgs(expected int, got []string) error {
	numArgs := len(got)
	if numArgs != expected {
		return fmt.Errorf(fmt.Sprintf("Expected %d arguments but only got %d (%v)", expected, numArgs, got))
	}
	return nil
}
