package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/wjase/diffyam/pkg/diff"
	"github.com/wjase/diffyam/pkg/report"
)

func main() {
	// numbPtr := flag.Int("numb", 42, "an int")
	// boolPtr := flag.Bool("fork", false, "a bool")
	// var outputfile string
	// flag.StringVar(&svar, "svar", "bar", "a string var")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `
diffyam - list the structured changes between two yaml files.
           Outputs a report of the changelog as a yaml file.

Syntax: diffyam  yamlfile1 yamlfile2


`)

		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()

	if len(args) < 2 {
		fmt.Fprintf(flag.CommandLine.Output(), "Error: Two args required\n")
		flag.Usage()
	}
	//fmt.Printf("%s %d %t \n", svar, *numbPtr, *boolPtr)
	fmt.Println("tail:", args)
	oldSpec := args[0]
	newSpec := args[1]

	changes, err := diff.CompareYamlFiles(oldSpec, newSpec)
	if err != nil {
		fmt.Printf("ERROR: %v", err)
		os.Exit(-1)
	}

	report.WriteChanges(changes, os.Stdout)

}
