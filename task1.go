package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	from := flag.String("from", "", "currency code e.g. \"USD\" ")
	to := flag.String("to", "", "currency code e.g. \"NGN\" ")
	amount := flag.Float64("amount", 1.0, "amount to convert to")

	flag.Parse()

	if *from == "" {
		fmt.Fprintln(os.Stderr, "error: --from is required")
		return
	}
	if *to == "" {
		fmt.Fprintln(os.Stderr, "error: --to is required")
		return
	}

	fmt.Printf("Converting %.2f %s to %s\n", *amount, *from, *to)

}
