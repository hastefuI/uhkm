package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/hastefuI/uhkm/cmd"
)

func main() {
	os.Exit(run())
}

func run() int {
	if err := cmd.Execute(); err != nil {
		if errors.Is(err, cmd.ErrLintFailure) {
			return 1
		}
		fmt.Fprintln(os.Stderr, "uhkm:", err)
		return 2
	}
	return 0
}
