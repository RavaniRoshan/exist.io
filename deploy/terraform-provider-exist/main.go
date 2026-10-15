package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	version = "1.0.0"
	commit  = "none"
)

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "prints provider version")
	flag.Parse()

	if showVersion {
		fmt.Printf("terraform-provider-exist %s (commit %s)\n", version, commit)
		os.Exit(0)
	}

	fmt.Println("exist.io official Terraform / OpenTofu provider plugin")
}
