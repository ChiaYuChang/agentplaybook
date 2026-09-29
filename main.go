package main

import (
	"fmt"
	"os"

	"github.com/ChiaYuChang/agentplaybook/internal/cli"
	"github.com/ChiaYuChang/agentplaybook/internal/version"
)

func main() {
	if err := cli.Execute(os.Args[1:], os.Stdout, os.Stderr, version.Release()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
