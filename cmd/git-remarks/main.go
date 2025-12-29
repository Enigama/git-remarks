package main

import (
	"os"

	"github.com/Enigama/git-remarks/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

