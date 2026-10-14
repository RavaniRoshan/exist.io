// Command exist provides the command-line interface for the exist.io Git forge.
// It supports managing repositories, pull requests, applications, and authentication.
package main

import (
	"fmt"
	"os"

	"exist.io/internal/cli"
)

func main() {
	c := cli.New()
	if err := c.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "exist: %v\n", err)
		os.Exit(1)
	}
}
