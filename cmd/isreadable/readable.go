package main

import (
	"os"

	"github.com/go-shiori/go-readability"
)

// Errors if stdin is not readable
func main() {
	if !readability.IsReadable(os.Stdin) {
		os.Exit(1)
	}
}
