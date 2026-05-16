package main

import (
	"fmt"
	"strings"
)

func repeat(word string, n int, sep string) string {
	var result strings.Builder
	
	for i := 0; i < n; i++{
		result.WriteString(word)
		if i < n-1{
		result.WriteString(sep)
		}
	}

	return result.String()

}

func main() {
	fmt.Println(repeat("Go", 3, " | "))
	fmt.Println(repeat("Ha", 4, "!"))
	fmt.Println(repeat("Yes", 1, "-") )
}
