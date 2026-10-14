package main

import (
	"fmt"
	"os"

	"exist.io/internal/cli"
)

func main() {
	c := cli.New()
	if err := c.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "origin: %v\n", err)
		os.Exit(1)
	}
}
