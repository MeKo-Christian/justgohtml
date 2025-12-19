// Package constants defines HTML5 specification constants.
package constants

// ForeignAttribute represents a foreign (namespaced) attribute adjustment.
type ForeignAttribute struct {
	Prefix       string // Attribute prefix (e.g., "xlink", "xml"), or empty string
	LocalName    string // Local name of the attribute
	NamespaceURL string // Namespace URL
}

// VoidElements are elements that have no closing tag.
var VoidElements = map[string]bool{
	"area":   true,
	"base":   true,
	"br":     true,
	"col":    true,
	"embed":  true,
	"hr":     true,
	"img":    true,
	"input":  true,
	"link":   true,
	"meta":   true,
	"param":  true,
	"source": true,
	"track":  true,
	"wbr":    true,
}

// RawTextElements are elements whose content is raw text.
var RawTextElements = map[string]bool{
	"script": true,
	"style":  true,
}

// EscapableRawTextElements are elements with escapable raw text.
var EscapableRawTextElements = map[string]bool{
	"textarea": true,
	"title":    true,
}

// SpecialElements are elements that require special parsing behavior.
// Per HTML5 spec, these elements affect the stack of open elements during tree construction.
var SpecialElements = map[string]bool{
	"address":    true,
	"applet":     true,
	"area":       true,
	"article":    true,
	"aside":      true,
	"base":       true,
	"basefont":   true,
	"bgsound":    true,
	"blockquote": true,
	"body":       true,
	"br":         true,
	"button":     true,
	"caption":    true,
	"center":     true,
	"col":        true,
	"colgroup":   true,
	"dd":         true,
	"details":    true,
	"dialog":     true,
	"dir":        true,
	"div":        true,
	"dl":         true,
	"dt":         true,
	"embed":      true,
	"fieldset":   true,
	"figcaption": true,
	"figure":     true,
	"footer":     true,
	"form":       true,
	"frame":      true,
	"frameset":   true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"head":       true,
	"header":     true,
	"hgroup":     true,
	"hr":         true,
	"html":       true,
	"iframe":     true,
	"img":        true,
	"input":      true,
	"keygen":     true,
	"li":         true,
	"link":       true,
	"listing":    true,
	"main":       true,
	"marquee":    true,
	"menu":       true,
	"menuitem":   true,
	"meta":       true,
	"nav":        true,
	"noembed":    true,
	"noframes":   true,
	"noscript":   true,
	"object":     true,
	"ol":         true,
	"p":          true,
	"param":      true,
	"plaintext":  true,
	"pre":        true,
	"script":     true,
	"search":     true,
	"section":    true,
	"select":     true,
	"source":     true,
	"style":      true,
	"summary":    true,
	"table":      true,
	"tbody":      true,
	"td":         true,
	"template":   true,
	"textarea":   true,
	"tfoot":      true,
	"th":         true,
	"thead":      true,
	"title":      true,
	"tr":         true,
	"track":      true,
	"ul":         true,
	"wbr":        true,
}

// FormattingElements are elements used for text formatting.
var FormattingElements = map[string]bool{
	"a":      true,
	"b":      true,
	"big":    true,
	"code":   true,
	"em":     true,
	"font":   true,
	"i":      true,
	"nobr":   true,
	"s":      true,
	"small":  true,
	"strike": true,
	"strong": true,
	"tt":     true,
	"u":      true,
}

// TableFosterTargets are elements that trigger foster parenting.
var TableFosterTargets = map[string]bool{
	"table": true,
	"tbody": true,
	"tfoot": true,
	"thead": true,
	"tr":    true,
}

// TableAllowedChildren are elements allowed as direct children of table elements.
var TableAllowedChildren = map[string]bool{
	"caption":  true,
	"colgroup": true,
	"tbody":    true,
	"tfoot":    true,
	"thead":    true,
	"tr":       true,
	"td":       true,
	"th":       true,
	"script":   true,
	"template": true,
	"style":    true,
}

// ImpliedEndTagElements are elements that can have implied end tags.
var ImpliedEndTagElements = map[string]bool{
	"dd":       true,
	"dt":       true,
	"li":       true,
	"optgroup": true,
	"option":   true,
	"p":        true,
	"rb":       true,
	"rp":       true,
	"rt":       true,
	"rtc":      true,
}

// ThoroughlyImpliedEndTagElements are elements for thorough implied end tags.
var ThoroughlyImpliedEndTagElements = map[string]bool{
	"caption":  true,
	"colgroup": true,
	"dd":       true,
	"dt":       true,
	"li":       true,
	"optgroup": true,
	"option":   true,
	"p":        true,
	"rb":       true,
	"rp":       true,
	"rt":       true,
	"rtc":      true,
	"tbody":    true,
	"td":       true,
	"tfoot":    true,
	"th":       true,
	"thead":    true,
	"tr":       true,
}

// SVGTagNameAdjustments maps lowercase SVG tag names to their proper camelCase form.
// Per HTML5 spec §13.2.6.5, SVG elements need case adjustment when parsed.
var SVGTagNameAdjustments = map[string]string{
	"altglyph":            "altGlyph",
	"altglyphdef":         "altGlyphDef",
	"altglyphitem":        "altGlyphItem",
	"animatecolor":        "animateColor",
	"animatemotion":       "animateMotion",
	"animatetransform":    "animateTransform",
	"clippath":            "clipPath",
	"feblend":             "feBlend",
	"fecolormatrix":       "feColorMatrix",
	"fecomponenttransfer": "feComponentTransfer",
	"fecomposite":         "feComposite",
	"feconvolvematrix":    "feConvolveMatrix",
	"fediffuselighting":   "feDiffuseLighting",
	"fedisplacementmap":   "feDisplacementMap",
	"fedistantlight":      "feDistantLight",
	"feflood":             "feFlood",
	"fefunca":             "feFuncA",
	"fefuncb":             "feFuncB",
	"fefuncg":             "feFuncG",
	"fefuncr":             "feFuncR",
	"fegaussianblur":      "feGaussianBlur",
	"feimage":             "feImage",
	"femerge":             "feMerge",
	"femergenode":         "feMergeNode",
	"femorphology":        "feMorphology",
	"feoffset":            "feOffset",
	"fepointlight":        "fePointLight",
	"fespecularlighting":  "feSpecularLighting",
	"fespotlight":         "feSpotLight",
	"fetile":              "feTile",
	"feturbulence":        "feTurbulence",
	"foreignobject":       "foreignObject",
	"glyphref":            "glyphRef",
	"lineargradient":      "linearGradient",
	"radialgradient":      "radialGradient",
	"textpath":            "textPath",
}

// SVGAttributeAdjustments maps lowercase SVG attribute names to their proper camelCase form.
// Per HTML5 spec §13.2.6.5, SVG attributes need case adjustment when parsed.
var SVGAttributeAdjustments = map[string]string{
	"attributename":       "attributeName",
	"attributetype":       "attributeType",
	"basefrequency":       "baseFrequency",
	"baseprofile":         "baseProfile",
	"calcmode":            "calcMode",
	"clippathunits":       "clipPathUnits",
	"diffuseconstant":     "diffuseConstant",
	"edgemode":            "edgeMode",
	"filterunits":         "filterUnits",
	"glyphref":            "glyphRef",
	"gradienttransform":   "gradientTransform",
	"gradientunits":       "gradientUnits",
	"kernelmatrix":        "kernelMatrix",
	"kernelunitlength":    "kernelUnitLength",
	"keypoints":           "keyPoints",
	"keysplines":          "keySplines",
	"keytimes":            "keyTimes",
	"lengthadjust":        "lengthAdjust",
	"limitingconeangle":   "limitingConeAngle",
	"markerheight":        "markerHeight",
	"markerunits":         "markerUnits",
	"markerwidth":         "markerWidth",
	"maskcontentunits":    "maskContentUnits",
	"maskunits":           "maskUnits",
	"numoctaves":          "numOctaves",
	"pathlength":          "pathLength",
	"patterncontentunits": "patternContentUnits",
	"patterntransform":    "patternTransform",
	"patternunits":        "patternUnits",
	"pointsatx":           "pointsAtX",
	"pointsaty":           "pointsAtY",
	"pointsatz":           "pointsAtZ",
	"preservealpha":       "preserveAlpha",
	"preserveaspectratio": "preserveAspectRatio",
	"primitiveunits":      "primitiveUnits",
	"refx":                "refX",
	"refy":                "refY",
	"repeatcount":         "repeatCount",
	"repeatdur":           "repeatDur",
	"requiredextensions":  "requiredExtensions",
	"requiredfeatures":    "requiredFeatures",
	"specularconstant":    "specularConstant",
	"specularexponent":    "specularExponent",
	"spreadmethod":        "spreadMethod",
	"startoffset":         "startOffset",
	"stddeviation":        "stdDeviation",
	"stitchtiles":         "stitchTiles",
	"surfacescale":        "surfaceScale",
	"systemlanguage":      "systemLanguage",
	"tablevalues":         "tableValues",
	"targetx":             "targetX",
	"targety":             "targetY",
	"textlength":          "textLength",
	"viewbox":             "viewBox",
	"viewtarget":          "viewTarget",
	"xchannelselector":    "xChannelSelector",
	"ychannelselector":    "yChannelSelector",
	"zoomandpan":          "zoomAndPan",
}

// MathMLAttributeAdjustments maps lowercase MathML attribute names to their proper camelCase form.
// Per HTML5 spec §13.2.6.5, MathML attributes need case adjustment when parsed.
var MathMLAttributeAdjustments = map[string]string{
	"definitionurl": "definitionURL",
}

// ForeignAttributeAdjustments maps lowercase attribute names to their namespaced form.
// Per HTML5 spec §13.2.6.5, foreign attributes need namespace adjustment when parsed.
var ForeignAttributeAdjustments = map[string]ForeignAttribute{
	"xlink:actuate": {Prefix: "xlink", LocalName: "actuate", NamespaceURL: "http://www.w3.org/1999/xlink"},
	"xlink:arcrole": {Prefix: "xlink", LocalName: "arcrole", NamespaceURL: "http://www.w3.org/1999/xlink"},
	"xlink:href":    {Prefix: "xlink", LocalName: "href", NamespaceURL: "http://www.w3.org/1999/xlink"},
	"xlink:role":    {Prefix: "xlink", LocalName: "role", NamespaceURL: "http://www.w3.org/1999/xlink"},
	"xlink:show":    {Prefix: "xlink", LocalName: "show", NamespaceURL: "http://www.w3.org/1999/xlink"},
	"xlink:title":   {Prefix: "xlink", LocalName: "title", NamespaceURL: "http://www.w3.org/1999/xlink"},
	"xlink:type":    {Prefix: "xlink", LocalName: "type", NamespaceURL: "http://www.w3.org/1999/xlink"},
	"xml:lang":      {Prefix: "xml", LocalName: "lang", NamespaceURL: "http://www.w3.org/XML/1998/namespace"},
	"xml:space":     {Prefix: "xml", LocalName: "space", NamespaceURL: "http://www.w3.org/XML/1998/namespace"},
	"xmlns":         {Prefix: "", LocalName: "xmlns", NamespaceURL: "http://www.w3.org/2000/xmlns/"},
	"xmlns:xlink":   {Prefix: "xmlns", LocalName: "xlink", NamespaceURL: "http://www.w3.org/2000/xmlns/"},
}

// Namespace URLs used in HTML5 parsing.
const (
	NamespaceHTML   = "http://www.w3.org/1999/xhtml"
	NamespaceSVG    = "http://www.w3.org/2000/svg"
	NamespaceMathML = "http://www.w3.org/1998/Math/MathML"
	NamespaceXLink  = "http://www.w3.org/1999/xlink"
	NamespaceXML    = "http://www.w3.org/XML/1998/namespace"
	NamespaceXMLNS  = "http://www.w3.org/2000/xmlns/"
)

// IntegrationPoint represents an element that serves as an integration point.
type IntegrationPoint struct {
	Namespace string
	LocalName string
}

// HTMLIntegrationPoints are SVG/MathML elements that allow HTML content.
// Per HTML5 spec §13.2.6.5, these elements switch back to HTML parsing mode.
var HTMLIntegrationPoints = map[IntegrationPoint]bool{
	{Namespace: NamespaceMathML, LocalName: "annotation-xml"}: true,
	{Namespace: NamespaceSVG, LocalName: "foreignObject"}:     true,
	{Namespace: NamespaceSVG, LocalName: "desc"}:              true,
	{Namespace: NamespaceSVG, LocalName: "title"}:             true,
}

// MathMLTextIntegrationPoints are MathML elements that allow text integration.
// Per HTML5 spec §13.2.6.5, these elements can contain text.
var MathMLTextIntegrationPoints = map[IntegrationPoint]bool{
	{Namespace: NamespaceMathML, LocalName: "mi"}:    true,
	{Namespace: NamespaceMathML, LocalName: "mo"}:    true,
	{Namespace: NamespaceMathML, LocalName: "mn"}:    true,
	{Namespace: NamespaceMathML, LocalName: "ms"}:    true,
	{Namespace: NamespaceMathML, LocalName: "mtext"}: true,
}

// ForeignBreakoutElements are HTML elements that break out of foreign content.
// Per HTML5 spec §13.2.6.5, these elements cause the parser to exit foreign content mode.
var ForeignBreakoutElements = map[string]bool{
	"b":          true,
	"big":        true,
	"blockquote": true,
	"body":       true,
	"br":         true,
	"center":     true,
	"code":       true,
	"dd":         true,
	"div":        true,
	"dl":         true,
	"dt":         true,
	"em":         true,
	"embed":      true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"head":       true,
	"hr":         true,
	"i":          true,
	"img":        true,
	"li":         true,
	"listing":    true,
	"menu":       true,
	"meta":       true,
	"nobr":       true,
	"ol":         true,
	"p":          true,
	"pre":        true,
	"ruby":       true,
	"s":          true,
	"small":      true,
	"span":       true,
	"strong":     true,
	"strike":     true,
	"sub":        true,
	"sup":        true,
	"table":      true,
	"tt":         true,
	"u":          true,
	"ul":         true,
	"var":        true,
}
