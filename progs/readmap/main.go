package main

import (
	// Directories in the root of the repo can be imported
	// as long as we pretend that they sit relative to the
	// url birc.au.dk/gsa, like this for the example 'shared':
	"fmt"
	"os"
	"strconv"

	"birc.au.dk/gsa/shared"
)

func printUsage(progname string) {
	fmt.Fprintf(os.Stderr,
		"Usage: %s -p genome\n       %s -d dist genome reads\n",
		progname, progname)
	os.Exit(1)
}

func main() {
	switch {
	case len(os.Args) == 3 && os.Args[1] == "-p":
		shared.Preprocess(os.Args[2])
	case len(os.Args) == 5 && os.Args[1] == "-d":
		dist, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting integer %s\n", os.Args[2])
			printUsage(os.Args[0])
		}
		genome := os.Args[3]
		reads := os.Args[4]
		shared.Readmap(genome, reads, dist)
	default:
		printUsage(os.Args[0])
	}
}
