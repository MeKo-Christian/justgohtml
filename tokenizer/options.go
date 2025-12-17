package tokenizer

// Options configure tokenizer behavior.
type Options struct {
	// DiscardBOM controls whether a leading U+FEFF BOM is removed from the input.
	// html5lib tokenizer tests set this per test case.
	DiscardBOM bool

	// XMLCoercion enables XML output coercions used by some test suites:
	// - U+000C FORM FEED becomes a space in text tokens
	// - Some non-XML characters become U+FFFD
	// - Comments replace "--" with "- -"
	XMLCoercion bool
}

func defaultOptions() Options {
	return Options{
		DiscardBOM: true,
	}
}
