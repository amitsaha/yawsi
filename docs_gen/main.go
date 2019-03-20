package main

import (
	"log"

	"github.com/spf13/cobra/doc"

	"github.com/amitsaha/yawsi/cmd"
)

func main() {

	// Markdown docs
	err := doc.GenMarkdownTree(cmd.RootCmd, "../docs/")
	if err != nil {
		log.Fatal(err)
	}

	header := &doc.GenManHeader{
		Title:   "MINE",
		Section: "3",
	}
	err = doc.GenManTree(cmd.RootCmd, header, "../docs/man_pages")
	if err != nil {
		log.Fatal(err)
	}
}
