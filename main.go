/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/
package main

import (
	"fmt"
	"memory/app"
	"memory/cmd"
	"os"
)

func main() {
	cliApp := cmd.CreateApp()
	err := cliApp.Run(os.Args)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	if err := app.Shutdown(); err != nil {
		fmt.Println("Error shutting down:", err)
	}
}
