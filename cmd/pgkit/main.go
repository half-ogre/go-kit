package main

import (
	"os"

	"github.com/half-ogre/go-kit/cmd/pgkit/subcmd"
)

func main() {
	if err := subcmd.Execute(); err != nil {
		os.Exit(1)
	}
}
