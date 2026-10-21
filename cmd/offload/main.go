// Command offload runs a heavy command in a free local heavy slot, else on the least-loaded host of
// an SSH pool with a free slot.
package main

import (
	"os"

	"github.com/eturkes/offload/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}
