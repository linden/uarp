package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/linden/uarp"
)

var output string

func init() {
	flag.StringVar(&output, "output", "", "directory to dump the table, will not dump if not provided")
	flag.Parse()
}

func main() {
	path := flag.Arg(0)

	if path == "" {
		panic("path is required")
	}

	raw, err := ioutil.ReadFile(path)

	if err != nil {
		panic(err)
	}

	fmt.Println("loaded", len(raw), "bytes", "\n")

	table := uarp.ParseTable(raw)

	fmt.Printf("version\n  %s\n", table.Version)

	fmt.Println("metadata")

	for _, meta := range table.Metadata {
		fmt.Printf("  type: %s, value: %d\n", meta.Type, meta.Value)
	}

	fmt.Println("rows")

	if output != "" {
		os.Mkdir(output, os.ModePerm)
	}

	for index, row := range table.Rows {
		fmt.Printf("  type: %s, size: %d, version: %s\n", row.Type, len(row.Payload), row.Version)

		for _, meta := range row.Metadata {
			fmt.Printf("    type: %s, size: %d\n", meta.Type, meta.Value)
		}

		if output != "" {
			err = ioutil.WriteFile(fmt.Sprintf("%s/%d", output, index), row.Payload, 0644)

			if err != nil {
				panic(err)
			}
		}
	}
}
