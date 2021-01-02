package main

import (
	"fmt"
	"os"
)

// docker         run image <cmd> <args>
// go run main.go run       <cmd> <args>

func main() {
	// println(os.Args)
	// println(os.Args[0])
	// println(len(os.Args))

	if len(os.Args) < 3 {
		fmt.Printf("\"run <cmd> [args]\" is required.\n")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "run":
		run()

	default:
		panic("unknown command")
	}
}

func run() {
	fmt.Printf("Running %v\n", os.Args[2:])
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
