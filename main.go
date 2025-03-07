package main

import (
	"log"

	"github.com/lpcalisi/otelcompare/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
