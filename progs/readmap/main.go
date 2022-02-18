package main

import (
	"fmt"
	"os"
	"strconv"

	"birc.au.dk/gsa"
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
		gsa.BwtPreproc(os.Args[2])
	case len(os.Args) == 5 && os.Args[1] == "-d":
		dist, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting integer %s\n", os.Args[2])
			printUsage(os.Args[0])
		}

		genomeFname := os.Args[3]
		readsFname := os.Args[4]
		genome := gsa.ReadPreprocTables(genomeFname)
		gsa.ScanFastq(readsFname, func(rec *gsa.FastqRecord) {
			for chrName, search := range genome {
				search(rec.Read, dist, func(i int32, cigar string) {
					gsa.PrintSam(rec.Name, chrName, i, cigar, rec.Read)
				})
			}
		})
	default:
		printUsage(os.Args[0])
	}
}
