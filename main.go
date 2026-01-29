package main

import (
	"os"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/cli"
)

func main() {
	err := cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}
