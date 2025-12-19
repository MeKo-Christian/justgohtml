// Command justgohtml is a CLI tool for parsing and querying HTML documents.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MeKo-Christian/JustGoHTML"
	"github.com/MeKo-Christian/JustGoHTML/dom"
	// Import selector package to register selector functions via init()
	_ "github.com/MeKo-Christian/JustGoHTML/selector"
	"github.com/MeKo-Christian/JustGoHTML/serialize"
)

// Output format constants.
const (
	outputFormatHTML     = "html"
	outputFormatText     = "text"
	outputFormatMarkdown = "markdown"
)

var version = "dev"

// config holds the CLI configuration.
type config struct {
	selector  string
	format    string
	first     bool
	separator string
	strip     bool
	pretty    bool
	indent    int
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	cfg, inputPath, err := parseFlags(args, stderr)
	if err != nil {
		return err
	}

	// Empty inputPath means version was shown
	if inputPath == "" {
		return nil
	}

	// Read input
	input, err := readInput(inputPath, stdin)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	// Parse HTML
	doc, err := JustGoHTML.ParseBytes(input)
	if err != nil {
		return fmt.Errorf("parsing HTML: %w", err)
	}

	// Get nodes to output
	var nodes []dom.Node
	if cfg.selector != "" {
		elements, err := doc.Query(cfg.selector)
		if err != nil {
			return fmt.Errorf("invalid selector: %w", err)
		}
		if cfg.first && len(elements) > 0 {
			elements = elements[:1]
		}
		for _, elem := range elements {
			nodes = append(nodes, elem)
		}
	} else {
		nodes = []dom.Node{doc}
	}

	// Format and output
	output := formatNodes(nodes, cfg)
	_, err = fmt.Fprint(stdout, output)
	return err
}

func parseFlags(args []string, stderr io.Writer) (*config, string, error) {
	fs := flag.NewFlagSet("justgohtml", flag.ContinueOnError)
	fs.SetOutput(stderr)

	cfg := &config{}

	// Define flags
	var selectorShort, formatShort string
	var showVersion, versionShort bool

	fs.StringVar(&cfg.selector, "selector", "", "CSS selector to filter output")
	fs.StringVar(&selectorShort, "s", "", "CSS selector to filter output (shorthand)")
	fs.StringVar(&cfg.format, "format", "html", "Output format: html, text, markdown")
	fs.StringVar(&formatShort, "f", "", "Output format (shorthand)")
	fs.BoolVar(&cfg.first, "first", false, "Output only first match")
	fs.StringVar(&cfg.separator, "separator", " ", "Separator for text output")
	fs.BoolVar(&cfg.strip, "strip", true, "Strip whitespace from text")
	fs.BoolVar(&cfg.pretty, "pretty", true, "Pretty-print HTML output")
	fs.IntVar(&cfg.indent, "indent", 2, "Indentation size for pretty-print")
	fs.BoolVar(&showVersion, "version", false, "Show version")
	fs.BoolVar(&versionShort, "v", false, "Show version (shorthand)")

	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: justgohtml [options] <file>\n\n")
		fmt.Fprintf(stderr, "Parse and query HTML documents.\n\n")
		fmt.Fprintf(stderr, "Arguments:\n")
		fmt.Fprintf(stderr, "  file    HTML file path or '-' for stdin\n\n")
		fmt.Fprintf(stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(stderr, "\nExamples:\n")
		fmt.Fprintf(stderr, "  justgohtml index.html                    Parse and pretty-print HTML\n")
		fmt.Fprintf(stderr, "  justgohtml -s 'p' index.html             Extract all <p> elements\n")
		fmt.Fprintf(stderr, "  justgohtml -s 'h1' -f text index.html    Extract h1 text content\n")
		fmt.Fprintf(stderr, "  curl -s URL | justgohtml -s 'title' -    Extract title from piped HTML\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil, "", nil
		}
		return nil, "", err
	}

	// Handle shorthand flags
	if selectorShort != "" && cfg.selector == "" {
		cfg.selector = selectorShort
	}
	if formatShort != "" && cfg.format == outputFormatHTML {
		cfg.format = formatShort
	}

	// Validate format
	switch cfg.format {
	case outputFormatHTML, outputFormatText, outputFormatMarkdown:
		// valid
	default:
		return nil, "", fmt.Errorf("invalid format %q: must be html, text, or markdown", cfg.format)
	}

	// Show version
	if showVersion || versionShort {
		fmt.Fprintf(stderr, "justgohtml version %s\n", version)
		return nil, "", nil
	}

	// Get input file
	remaining := fs.Args()
	if len(remaining) == 0 {
		fs.Usage()
		return nil, "", fmt.Errorf("missing input file")
	}

	return cfg, remaining[0], nil
}

func readInput(path string, stdin io.Reader) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(stdin)
	}
	return os.ReadFile(path)
}

func formatNodes(nodes []dom.Node, cfg *config) string {
	if len(nodes) == 0 {
		return ""
	}

	var results []string

	for _, node := range nodes {
		var result string
		switch cfg.format {
		case outputFormatHTML:
			result = formatHTML(node, cfg)
		case outputFormatText:
			result = formatText(node, cfg)
		case outputFormatMarkdown:
			result = formatMarkdown(node, cfg)
		}
		if result != "" {
			results = append(results, result)
		}
	}

	output := strings.Join(results, "\n")
	if output != "" && !strings.HasSuffix(output, "\n") {
		output += "\n"
	}
	return output
}

func formatHTML(node dom.Node, cfg *config) string {
	opts := serialize.Options{
		Pretty:     cfg.pretty,
		IndentSize: cfg.indent,
	}
	return serialize.ToHTML(node, opts)
}

func formatText(node dom.Node, cfg *config) string {
	text := extractText(node)
	if cfg.strip {
		text = collapseWhitespace(text)
	}
	return text
}

func formatMarkdown(node dom.Node, _ *config) string {
	return toMarkdown(node)
}

// extractText extracts all text content from a node.
func extractText(node dom.Node) string {
	var sb strings.Builder
	extractTextRecursive(node, &sb)
	return sb.String()
}

func extractTextRecursive(node dom.Node, sb *strings.Builder) {
	switch n := node.(type) {
	case *dom.Text:
		sb.WriteString(n.Data)
	case *dom.Element:
		for _, child := range n.Children() {
			extractTextRecursive(child, sb)
		}
	case *dom.Document:
		for _, child := range n.Children() {
			extractTextRecursive(child, sb)
		}
	}
}

// collapseWhitespace collapses runs of whitespace into single spaces and trims.
func collapseWhitespace(s string) string {
	var sb strings.Builder
	inWhitespace := true // Start true to trim leading whitespace
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f' {
			if !inWhitespace {
				sb.WriteByte(' ')
				inWhitespace = true
			}
		} else {
			sb.WriteRune(r)
			inWhitespace = false
		}
	}
	result := sb.String()
	// Trim trailing space
	if len(result) > 0 && result[len(result)-1] == ' ' {
		result = result[:len(result)-1]
	}
	return result
}

// toMarkdown converts a node to Markdown format.
func toMarkdown(node dom.Node) string {
	var sb strings.Builder
	toMarkdownRecursive(node, &sb, 0)
	return strings.TrimSpace(sb.String())
}

func toMarkdownRecursive(node dom.Node, sb *strings.Builder, listDepth int) {
	switch n := node.(type) {
	case *dom.Text:
		text := collapseWhitespace(n.Data)
		if text != "" {
			sb.WriteString(text)
		}
	case *dom.Element:
		mdElementToMarkdown(n, sb, listDepth)
	case *dom.Document:
		for _, child := range n.Children() {
			toMarkdownRecursive(child, sb, listDepth)
		}
	}
}

// mdElementToMarkdown converts an HTML element to Markdown.
func mdElementToMarkdown(n *dom.Element, sb *strings.Builder, listDepth int) {
	switch n.TagName {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		mdWriteHeading(n, sb)
	case "p":
		mdWriteParagraph(n, sb, listDepth)
	case "br":
		sb.WriteString("  \n")
	case "hr":
		sb.WriteString("\n---\n\n")
	case "strong", "b":
		mdWriteInlineFormatted(n, sb, listDepth, "**")
	case "em", "i":
		mdWriteInlineFormatted(n, sb, listDepth, "*")
	case "code":
		sb.WriteString("`")
		writeChildrenText(n, sb)
		sb.WriteString("`")
	case "pre":
		sb.WriteString("```\n")
		writeChildrenText(n, sb)
		sb.WriteString("\n```\n\n")
	case "a":
		mdWriteLink(n, sb)
	case "img":
		mdWriteImage(n, sb)
	case "ul":
		mdWriteUnorderedList(n, sb, listDepth)
	case "ol":
		mdWriteOrderedList(n, sb, listDepth)
	case "blockquote":
		mdWriteBlockquote(n, sb)
	case "table":
		writeTable(n, sb)
	case "script", "style", "head":
		// Skip these elements
	default:
		for _, child := range n.Children() {
			toMarkdownRecursive(child, sb, listDepth)
		}
	}
}

func mdWriteHeading(n *dom.Element, sb *strings.Builder) {
	level := int(n.TagName[1] - '0')
	sb.WriteString(strings.Repeat("#", level))
	sb.WriteString(" ")
	writeChildrenText(n, sb)
	sb.WriteString("\n\n")
}

func mdWriteParagraph(n *dom.Element, sb *strings.Builder, listDepth int) {
	for _, child := range n.Children() {
		toMarkdownRecursive(child, sb, listDepth)
	}
	sb.WriteString("\n\n")
}

func mdWriteInlineFormatted(n *dom.Element, sb *strings.Builder, listDepth int, marker string) {
	sb.WriteString(marker)
	for _, child := range n.Children() {
		toMarkdownRecursive(child, sb, listDepth)
	}
	sb.WriteString(marker)
}

func mdWriteLink(n *dom.Element, sb *strings.Builder) {
	href := n.Attr("href")
	sb.WriteString("[")
	writeChildrenText(n, sb)
	sb.WriteString("](")
	sb.WriteString(href)
	sb.WriteString(")")
}

func mdWriteImage(n *dom.Element, sb *strings.Builder) {
	alt := n.Attr("alt")
	src := n.Attr("src")
	sb.WriteString("![")
	sb.WriteString(alt)
	sb.WriteString("](")
	sb.WriteString(src)
	sb.WriteString(")")
}

func mdWriteUnorderedList(n *dom.Element, sb *strings.Builder, listDepth int) {
	sb.WriteString("\n")
	for _, child := range n.Children() {
		if elem, ok := child.(*dom.Element); ok && elem.TagName == "li" {
			sb.WriteString(strings.Repeat("  ", listDepth))
			sb.WriteString("- ")
			for _, liChild := range elem.Children() {
				toMarkdownRecursive(liChild, sb, listDepth+1)
			}
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")
}

func mdWriteOrderedList(n *dom.Element, sb *strings.Builder, listDepth int) {
	sb.WriteString("\n")
	num := 1
	for _, child := range n.Children() {
		if elem, ok := child.(*dom.Element); ok && elem.TagName == "li" {
			sb.WriteString(strings.Repeat("  ", listDepth))
			fmt.Fprintf(sb, "%d. ", num)
			for _, liChild := range elem.Children() {
				toMarkdownRecursive(liChild, sb, listDepth+1)
			}
			sb.WriteString("\n")
			num++
		}
	}
	sb.WriteString("\n")
}

func mdWriteBlockquote(n *dom.Element, sb *strings.Builder) {
	lines := strings.Split(extractText(n), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			sb.WriteString("> ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")
}

func writeChildrenText(elem *dom.Element, sb *strings.Builder) {
	text := extractText(elem)
	text = collapseWhitespace(text)
	sb.WriteString(text)
}

func writeTable(table *dom.Element, sb *strings.Builder) {
	headers, rows := extractTableData(table)

	if len(headers) == 0 && len(rows) == 0 {
		return
	}

	colCount := normalizeTableData(&headers, rows)
	writeMarkdownTable(sb, headers, rows, colCount)
}

func extractTableData(table *dom.Element) ([]string, [][]string) {
	var headers []string
	var rows [][]string

	for _, child := range table.Children() {
		elem, ok := child.(*dom.Element)
		if !ok {
			continue
		}

		switch elem.TagName {
		case "thead":
			headers = extractTableHeader(elem)
		case "tbody":
			rows = append(rows, extractTableBodyRows(elem)...)
		case "tr":
			headers, rows = handleDirectTableRow(elem, headers, rows)
		}
	}
	return headers, rows
}

func extractTableHeader(thead *dom.Element) []string {
	for _, tr := range thead.Children() {
		if trElem, ok := tr.(*dom.Element); ok && trElem.TagName == "tr" {
			headers := extractTableRow(trElem, "th")
			if len(headers) == 0 {
				headers = extractTableRow(trElem, "td")
			}
			return headers
		}
	}
	return nil
}

func extractTableBodyRows(tbody *dom.Element) [][]string {
	var rows [][]string
	for _, tr := range tbody.Children() {
		if trElem, ok := tr.(*dom.Element); ok && trElem.TagName == "tr" {
			row := extractTableRow(trElem, "td")
			if len(row) > 0 {
				rows = append(rows, row)
			}
		}
	}
	return rows
}

func handleDirectTableRow(elem *dom.Element, headers []string, rows [][]string) ([]string, [][]string) {
	cells := extractTableRow(elem, "th")
	if len(cells) > 0 && len(headers) == 0 {
		return cells, rows
	}
	cells = extractTableRow(elem, "td")
	if len(cells) > 0 {
		rows = append(rows, cells)
	}
	return headers, rows
}

func normalizeTableData(headers *[]string, rows [][]string) int {
	colCount := len(*headers)
	for _, row := range rows {
		if len(row) > colCount {
			colCount = len(row)
		}
	}

	for len(*headers) < colCount {
		*headers = append(*headers, "")
	}
	for i := range rows {
		for len(rows[i]) < colCount {
			rows[i] = append(rows[i], "")
		}
	}
	return colCount
}

func writeMarkdownTable(sb *strings.Builder, headers []string, rows [][]string, colCount int) {
	sb.WriteString("| ")
	sb.WriteString(strings.Join(headers, " | "))
	sb.WriteString(" |\n")

	sb.WriteString("|")
	for range colCount {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")

	for _, row := range rows {
		sb.WriteString("| ")
		sb.WriteString(strings.Join(row, " | "))
		sb.WriteString(" |\n")
	}
	sb.WriteString("\n")
}

func extractTableRow(tr *dom.Element, cellTag string) []string {
	var cells []string
	for _, child := range tr.Children() {
		if elem, ok := child.(*dom.Element); ok && elem.TagName == cellTag {
			text := collapseWhitespace(extractText(elem))
			cells = append(cells, text)
		}
	}
	return cells
}
