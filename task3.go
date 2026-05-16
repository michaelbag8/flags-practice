package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	input := flag.String("input", "", "filename to read")
	output := flag.String("output", "", "filename to write")
	uppercase := flag.Bool("uppercase", false, "convert to uppercase")
	count := flag.Bool("count", false, "prints word count")

	flag.Parse()

	data, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: reading file")
		return
	}

	content := string(data)
	if *uppercase {
		content = strings.ToUpper(content)
	}

	if *count {
		words := strings.Fields(content)
		fmt.Printf("word count: %d\n", len(words))
		return
	}

	if *output != "" {
		err = os.WriteFile(*output, []byte(content), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: writing file")
			return
		}
	} else {
		fmt.Println(content)
	}
}
