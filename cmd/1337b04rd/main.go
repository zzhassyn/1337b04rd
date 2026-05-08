package main

import (
	"1337b04rd/internal/app"
	"fmt"
	"os"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
