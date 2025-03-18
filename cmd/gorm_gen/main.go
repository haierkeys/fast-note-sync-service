package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/haierkeys/obsidian-better-sync-service/cmd/gorm_gen/pkg"
)

var (
	input   string
	structs []string
	prefix  string
)

func init() {
	flagStructs := flag.String("structs", "", "[Required] The name of schema structs to generate structs for, comma seperated\n")
	flagInput := flag.String("input", "", "[Required] The name of the input file dir\n")
	flagpre := flag.String("pre", "", "[Required] db_driver.TablePrefix\n")
	flag.Parse()

	if *flagStructs == "" || *flagInput == "" {
		flag.Usage()
		os.Exit(1)
	}

	structs = strings.Split(*flagStructs, ",")
	input = *flagInput
	prefix = *flagpre
}

func main() {
	gen := pkg.NewGenerator(input)
	p := pkg.NewParser(input)
	if err := gen.ParserAST(p, structs, prefix).Generate().Format().Flush(); err != nil {
		log.Fatalln(err)
	}
}
