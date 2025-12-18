// Command justhtml is a CLI tool for parsing and querying HTML documents.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/MeKo-Christian/JustGoHTML"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Define flags
	selector := flag.String("selector", "", "CSS selector to filter output")
	selectorShort := flag.String("s", "", "CSS selector to filter output (shorthand)")
	format := flag.String("format", "html", "Output format: html, text, markdown")
	formatShort := flag.String("f", "", "Output format (shorthand)")
	first := flag.Bool("first", false, "Output only first match")
	separator := flag.String("separator", " ", "Separator for text output")
	strip := flag.Bool("strip", true, "Strip whitespace from text")
	pretty := flag.Bool("pretty", true, "Pretty-print HTML output")
	indent := flag.Int("indent", 2, "Indentation size for pretty-print")
	showVersion := flag.Bool("version", false, "Show version")
	versionShort := flag.Bool("v", false, "Show version (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <file>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Parse and query HTML documents.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  file    HTML file path or '-' for stdin\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Handle shorthand flags
	if *selectorShort != "" && *selector == "" {
		*selector = *selectorShort
	}
	if *formatShort != "" && *format == "html" {
		*format = *formatShort
	}

	// Show version
	if *showVersion || *versionShort {
		fmt.Printf("justhtml version %s\n", version)
		return nil
	}

	// Get input file
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		return fmt.Errorf("missing input file")
	}

	inputPath := args[0]

	// Read input
	var input []byte
	var err error

	if inputPath == "-" {
		input, err = io.ReadAll(os.Stdin)
	} else {
		input, err = os.ReadFile(inputPath)
	}
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	// Parse HTML
	doc, err := JustGoHTML.ParseBytes(input)
	if err != nil {
		return fmt.Errorf("parsing HTML: %w", err)
	}

	// For now, just show that parsing worked
	_ = doc
	_ = selector
	_ = format
	_ = first
	_ = separator
	_ = strip
	_ = pretty
	_ = indent

	fmt.Println("TODO: Implement output formatting")
	return nil
}
