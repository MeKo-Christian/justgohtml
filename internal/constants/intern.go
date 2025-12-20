package constants

// Package-level string interning for common HTML tag and attribute names.
// This reduces memory allocations during parsing by reusing pre-allocated strings.

// CommonTagNames contains the most frequently used HTML tag names.
// These are pre-allocated to avoid repeated string allocations during tokenization.
var CommonTagNames = map[string]string{
	// Document structure
	"html":  "html",
	"head":  "head",
	"body":  "body",
	"title": "title",
	"meta":  "meta",
	"link":  "link",
	"style": "style",

	// Sectioning
	"header":  "header",
	"footer":  "footer",
	"nav":     "nav",
	"section": "section",
	"article": "article",
	"aside":   "aside",
	"main":    "main",

	// Text content
	"div":        "div",
	"p":          "p",
	"span":       "span",
	"h1":         "h1",
	"h2":         "h2",
	"h3":         "h3",
	"h4":         "h4",
	"h5":         "h5",
	"h6":         "h6",
	"blockquote": "blockquote",
	"pre":        "pre",
	"code":       "code",

	// Lists
	"ul": "ul",
	"ol": "ol",
	"li": "li",
	"dl": "dl",
	"dt": "dt",
	"dd": "dd",

	// Tables
	"table":    "table",
	"thead":    "thead",
	"tbody":    "tbody",
	"tfoot":    "tfoot",
	"tr":       "tr",
	"th":       "th",
	"td":       "td",
	"caption":  "caption",
	"colgroup": "colgroup",
	"col":      "col",

	// Forms
	"form":     "form",
	"input":    "input",
	"button":   "button",
	"select":   "select",
	"option":   "option",
	"textarea": "textarea",
	"label":    "label",
	"fieldset": "fieldset",
	"legend":   "legend",

	// Media
	"img":    "img",
	"video":  "video",
	"audio":  "audio",
	"source": "source",
	"track":  "track",
	"canvas": "canvas",
	"svg":    "svg",

	// Interactive
	"a":        "a",
	"script":   "script",
	"noscript": "noscript",
	"iframe":   "iframe",

	// Text formatting
	"b":      "b",
	"i":      "i",
	"u":      "u",
	"s":      "s",
	"em":     "em",
	"strong": "strong",
	"small":  "small",
	"mark":   "mark",
	"del":    "del",
	"ins":    "ins",
	"sub":    "sub",
	"sup":    "sup",

	// Other common elements
	"br":       "br",
	"hr":       "hr",
	"template": "template",
	"slot":     "slot",
	"base":     "base",
}

// CommonAttributeNames contains the most frequently used HTML attribute names.
// These are pre-allocated to avoid repeated string allocations during tokenization.
var CommonAttributeNames = map[string]string{
	// Global attributes
	"id":    "id",
	"class": "class",
	"style": "style",
	"title": "title",
	"lang":  "lang",
	"dir":   "dir",

	// Data attributes (common patterns)
	"data-id":    "data-id",
	"data-name":  "data-name",
	"data-value": "data-value",

	// Link attributes
	"href":   "href",
	"rel":    "rel",
	"target": "target",
	"type":   "type",

	// Media attributes
	"src":    "src",
	"alt":    "alt",
	"width":  "width",
	"height": "height",

	// Form attributes
	"name":        "name",
	"value":       "value",
	"placeholder": "placeholder",
	"disabled":    "disabled",
	"readonly":    "readonly",
	"required":    "required",
	"checked":     "checked",
	"selected":    "selected",
	"action":      "action",
	"method":      "method",
	"for":         "for",

	// Interactive attributes
	"onclick":    "onclick",
	"onchange":   "onchange",
	"onsubmit":   "onsubmit",
	"onload":     "onload",
	"tabindex":   "tabindex",
	"aria-label": "aria-label",
	"role":       "role",

	// Meta attributes
	"content":  "content",
	"charset":  "charset",
	"property": "property",

	// Other common attributes
	"hidden":       "hidden",
	"data":         "data",
	"download":     "download",
	"enctype":      "enctype",
	"accept":       "accept",
	"autocomplete": "autocomplete",
	"autofocus":    "autofocus",
	"maxlength":    "maxlength",
	"minlength":    "minlength",
	"pattern":      "pattern",
	"multiple":     "multiple",
	"size":         "size",
	"min":          "min",
	"max":          "max",
	"step":         "step",
	"colspan":      "colspan",
	"rowspan":      "rowspan",
	"scope":        "scope",
	"headers":      "headers",
}

// InternTagName returns an interned version of the tag name if it's a common tag,
// otherwise returns the original string.
func InternTagName(name string) string {
	if interned, ok := CommonTagNames[name]; ok {
		return interned
	}
	return name
}

// InternAttributeName returns an interned version of the attribute name if it's a common attribute,
// otherwise returns the original string.
func InternAttributeName(name string) string {
	if interned, ok := CommonAttributeNames[name]; ok {
		return interned
	}
	return name
}
