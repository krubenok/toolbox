package main

import (
	"os"

	"github.com/krubenok/toolbox/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
