package main

import "fmt"

// show can return customer implementation of error type
type MyError struct {
	Message string
}

func (e MyError) Error() string { return fmt.Sprintf("ERROR: %s!!!", e.Message) }
