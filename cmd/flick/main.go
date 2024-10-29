package main

import (
	"github.com/Wraient/flick/internal"
	"fmt"
)


// Example usage:
func main() {
    internal.ClearScreen()
	var query string
	fmt.Print("Enter search query: ")
	fmt.Scanln(&query)

	results, err := internal.SearchMedia(query)
	if err != nil {
		internal.ExitFlick(err)
	}

    fmt.Printf("Results: %v\n", results)
	// The selection and playback is handled within SearchMedia
}