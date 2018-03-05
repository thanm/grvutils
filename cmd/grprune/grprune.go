package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/thanm/grvutils/grparser"
	"github.com/thanm/grvutils/grprune"
	"github.com/thanm/grvutils/zgr"
)

var verbflag = flag.Int("v", 0, "Verbose trace output level")
var depthflag = flag.Int("d", 4, "Prune depth from root.")
var modeflag = flag.String("m", "both", "Prune mode. One of {fwd,bwd,both}.")
var infileflag = flag.String("i", "", "Input file")
var outfileflag = flag.String("o", "", "Output file")
var rootidflag = flag.String("r", "", "Root node ID")

func verb(vlevel int, s string, a ...interface{}) {
	if *verbflag >= vlevel {
		fmt.Printf(s, a...)
		fmt.Printf("\n")
	}
}

func usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: grprune [flags]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("grprune: ")
	flag.Parse()
	verb(1, "in main")
	if flag.NArg() != 0 {
		usage("unknown extra args")
	}
	if *rootidflag == "" {
		usage("specify root node ID with -r flag")
	}
	if *modeflag != "both" && *modeflag != "fwd" && *modeflag != "bwd" {
		usage(fmt.Sprintf("illegal mode '%s'", *modeflag))
	}
	var err error
	var infile *os.File = os.Stdin
	if len(*infileflag) > 0 {
		verb(1, "opening %s", *infileflag)
		infile, err = os.Open(*infileflag)
		if err != nil {
			log.Fatal(err)
		}
	}
	var outfile *os.File = os.Stdout
	if len(*outfileflag) > 0 {
		verb(1, "opening %s", *outfileflag)
		outfile, err = os.OpenFile(*outfileflag, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}
	g := zgr.NewGraph()
	err = grparser.ParseGraph(infile, g)
	if err != nil {
		log.Fatal(err)
	}
	err = grprune.PruneGraph(g, *rootidflag, *modeflag, *depthflag, outfile)
	if err != nil {
		log.Fatal(err)
	}
	verb(1, "leaving main")
}
