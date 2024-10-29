package main

import (
	"github.com/Wraient/flick/internal"
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Example usage:
func main() {
    // internal.ClearScreen()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter search query: ")
	query, _ := reader.ReadString('\n')
	query = strings.TrimSpace(query) // Remove any trailing newline

	results := internal.SearchMedia(query)
	if results == nil {
		internal.ExitFlick(fmt.Errorf("no results found"))
	}

    fmt.Printf("Results: %v\n", results)
}