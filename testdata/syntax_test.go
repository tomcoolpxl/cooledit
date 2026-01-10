// This is a test file for syntax highlighting verification
// Open this file in cooledit to verify syntax highlighting works

package main

import (
	"fmt"
	"strings"
)

// Constants should be highlighted
const (
	MaxSize   = 100
	Version   = "1.0.0"
	IsEnabled = true
)

// Types should be highlighted
type Person struct {
	Name string
	Age  int
}

// Functions should be highlighted
func main() {
	// Keywords: func, if, else, for, return, var, const
	var name = "World"

	// Strings should be highlighted
	greeting := fmt.Sprintf("Hello, %s!", name)
	fmt.Println(greeting)

	// Numbers should be highlighted
	count := 42
	pi := 3.14159
	hex := 0xFF

	// Comments should be highlighted (like this one)

	// Operators: + - * / = == != < > && ||
	if count > 10 && pi < 4.0 {
		fmt.Println("Math works!")
	}

	// Builtin functions
	length := len(name)
	result := make([]int, length)

	// For loop with range
	for i, v := range result {
		result[i] = v + 1
	}

	// String operations
	upper := strings.ToUpper(name)
	fmt.Println(upper)
}

/* Multi-line comment
   should also be highlighted
   as comments */
