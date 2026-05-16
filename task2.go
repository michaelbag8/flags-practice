package main

import (
	"flag"
	"fmt"
)

func main() {

	name := flag.String("name", "Go", "name of the user")
	times := flag.Int("times", 3, "number of repeats")

	flag.Parse()
	for i := 0; i < *times; i++ {
		fmt.Println(*name)
	}

}
