package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/Lokee86/grimoire/internal/app"
)

func main() {
	if err := app.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		var exitError *app.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.Code)
		}
		fmt.Fprintln(os.Stderr, "grimoire:", err)
		os.Exit(1)
	}
}
