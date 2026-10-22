// Command pooladdrs prints each pool host's address, one per line, for tools/install: the pool
// file read by offload's own parser (P1, P2), so every file offload accepts installs everywhere.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/eturkes/offload/internal/pool"
)

func main() {
	if err := run(os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "pooladdrs: %v\n", err)
		os.Exit(1)
	}
}

func run(getenv func(string) string, stdout io.Writer) error {
	hosts, err := pool.Load(pool.Path(getenv))
	if err != nil {
		return err
	}
	for _, h := range hosts {
		fmt.Fprintln(stdout, h.Address)
	}
	return nil
}
