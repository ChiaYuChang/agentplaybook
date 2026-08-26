package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/ChiaYuChang/workflow/internal/cli"
)

var version = detectVersion()

func main() {
	if err := cli.Execute(os.Args[1:], os.Stdout, os.Stderr, version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func detectVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "dev"
	}
	return info.Main.Version
}
