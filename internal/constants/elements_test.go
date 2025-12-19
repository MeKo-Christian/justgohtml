package constants

import (
	"testing"
)

func TestSpecialElements(t *testing.T) {
	// Test that all expected special elements are present
	expectedSpecial := []string{
		"address", "applet", "area", "article", "aside", "base", "basefont", "bgsound",
		"blockquote", "body", "br", "button", "caption", "center", "col", "colgroup",
		"dd", "details", "dialog", "dir", "div", "dl", "dt", "embed", "fieldset",
		"figcaption", "figure", "footer", "form", "frame", "frameset", "h1", "h2",
		"h3", "h4", "h5", "h6", "head", "header", "hgroup", "hr", "html", "iframe",
		"img", "input", "keygen", "li", "link", "listing", "main", "marquee", "menu",
		"menuitem", "meta", "nav", "noembed", "noframes", "noscript", "object", "ol",
		"p", "param", "plaintext", "pre", "script", "search", "section", "select",
		"source", "style", "summary", "table", "tbody", "td", "template", "textarea",
		"tfoot", "th", "thead", "title", "tr", "track", "ul", "wbr",
	}

	for _, elem := range expectedSpecial {
		t.Run(elem, func(t *testing.T) {
			if !SpecialElements[elem] {
				t.Errorf("SpecialElements[%q] = false, want true", elem)
			}
		})
	}

	// Test that the count is correct
	if len(SpecialElements) != len(expectedSpecial) {
		t.Errorf("SpecialElements has %d entries, want %d", len(SpecialElements), len(expectedSpecial))
	}

	// Test that xmp is NOT in special elements (common mistake)
	if SpecialElements["xmp"] {
		t.Error("SpecialElements[\"xmp\"] = true, should not be a special element")
	}
}

func TestFormattingElements(t *testing.T) {
	expectedFormatting := []string{
		"a", "b", "big", "code", "em", "font", "i", "nobr", "s", "small",
		"strike", "strong", "tt", "u",
	}

	for _, elem := range expectedFormatting {
		t.Run(elem, func(t *testing.T) {
			if !FormattingElements[elem] {
				t.Errorf("FormattingElements[%q] = false, want true", elem)
			}
		})
	}

	if len(FormattingElements) != len(expectedFormatting) {
		t.Errorf("FormattingElements has %d entries, want %d", len(FormattingElements), len(expectedFormatting))
	}
}

func TestVoidElements(t *testing.T) {
	expectedVoid := []string{
		"area", "base", "br", "col", "embed", "hr", "img", "input", "keygen",
		"link", "meta", "param", "source", "track", "wbr",
	}

	for _, elem := range expectedVoid {
		t.Run(elem, func(t *testing.T) {
			if !VoidElements[elem] {
				t.Errorf("VoidElements[%q] = false, want true", elem)
			}
		})
	}

	if len(VoidElements) != len(expectedVoid) {
		t.Errorf("VoidElements has %d entries, want %d", len(VoidElements), len(expectedVoid))
	}

	// Test that non-void elements are not in the set
	nonVoid := []string{"div", "span", "p", "script", "style"}
	for _, elem := range nonVoid {
		t.Run("not-"+elem, func(t *testing.T) {
			if VoidElements[elem] {
				t.Errorf("VoidElements[%q] = true, want false", elem)
			}
		})
	}
}

func TestRawTextElements(t *testing.T) {
	expectedRawText := []string{"script", "style"}

	for _, elem := range expectedRawText {
		t.Run(elem, func(t *testing.T) {
			if !RawTextElements[elem] {
				t.Errorf("RawTextElements[%q] = false, want true", elem)
			}
		})
	}

	if len(RawTextElements) != len(expectedRawText) {
		t.Errorf("RawTextElements has %d entries, want %d", len(RawTextElements), len(expectedRawText))
	}
}

func TestEscapableRawTextElements(t *testing.T) {
	expectedEscapable := []string{"textarea", "title"}

	for _, elem := range expectedEscapable {
		t.Run(elem, func(t *testing.T) {
			if !EscapableRawTextElements[elem] {
				t.Errorf("EscapableRawTextElements[%q] = false, want true", elem)
			}
		})
	}

	if len(EscapableRawTextElements) != len(expectedEscapable) {
		t.Errorf("EscapableRawTextElements has %d entries, want %d", len(EscapableRawTextElements), len(expectedEscapable))
	}
}

func TestImpliedEndTagElements(t *testing.T) {
	expectedImplied := []string{
		"dd", "dt", "li", "optgroup", "option", "p", "rb", "rp", "rt", "rtc",
	}

	for _, elem := range expectedImplied {
		t.Run(elem, func(t *testing.T) {
			if !ImpliedEndTagElements[elem] {
				t.Errorf("ImpliedEndTagElements[%q] = false, want true", elem)
			}
		})
	}

	if len(ImpliedEndTagElements) != len(expectedImplied) {
		t.Errorf("ImpliedEndTagElements has %d entries, want %d", len(ImpliedEndTagElements), len(expectedImplied))
	}
}

func TestThoroughlyImpliedEndTagElements(t *testing.T) {
	expectedThoroughly := []string{
		"caption", "colgroup", "dd", "dt", "li", "optgroup", "option", "p",
		"rb", "rp", "rt", "rtc", "tbody", "td", "tfoot", "th", "thead", "tr",
	}

	for _, elem := range expectedThoroughly {
		t.Run(elem, func(t *testing.T) {
			if !ThoroughlyImpliedEndTagElements[elem] {
				t.Errorf("ThoroughlyImpliedEndTagElements[%q] = false, want true", elem)
			}
		})
	}

	if len(ThoroughlyImpliedEndTagElements) != len(expectedThoroughly) {
		t.Errorf("ThoroughlyImpliedEndTagElements has %d entries, want %d",
			len(ThoroughlyImpliedEndTagElements), len(expectedThoroughly))
	}

	// Test that thoroughly implied includes all regularly implied
	for elem := range ImpliedEndTagElements {
		if !ThoroughlyImpliedEndTagElements[elem] {
			t.Errorf("Element %q is in ImpliedEndTagElements but not in ThoroughlyImpliedEndTagElements", elem)
		}
	}
}

func TestElementSetConsistency(t *testing.T) {
	// Void elements should be special elements
	for elem := range VoidElements {
		if !SpecialElements[elem] {
			t.Errorf("Void element %q is not in SpecialElements", elem)
		}
	}

	// Raw text elements should be special elements
	for elem := range RawTextElements {
		if !SpecialElements[elem] {
			t.Errorf("RawText element %q is not in SpecialElements", elem)
		}
	}

	// Escapable raw text elements should be special elements
	for elem := range EscapableRawTextElements {
		if !SpecialElements[elem] {
			t.Errorf("EscapableRawText element %q is not in SpecialElements", elem)
		}
	}
}

func TestSpecialElementsDialogAndMenuitem(t *testing.T) {
	// Regression test: ensure dialog and menuitem are included
	if !SpecialElements["dialog"] {
		t.Error("dialog should be in SpecialElements")
	}
	if !SpecialElements["menuitem"] {
		t.Error("menuitem should be in SpecialElements")
	}

	// Regression test: ensure xmp is NOT included
	if SpecialElements["xmp"] {
		t.Error("xmp should not be in SpecialElements")
	}
}

func TestSVGTagNameAdjustments(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"altglyph", "altGlyph"},
		{"altglyphdef", "altGlyphDef"},
		{"altglyphitem", "altGlyphItem"},
		{"animatecolor", "animateColor"},
		{"animatemotion", "animateMotion"},
		{"animatetransform", "animateTransform"},
		{"clippath", "clipPath"},
		{"feblend", "feBlend"},
		{"fecolormatrix", "feColorMatrix"},
		{"fecomponenttransfer", "feComponentTransfer"},
		{"fecomposite", "feComposite"},
		{"feconvolvematrix", "feConvolveMatrix"},
		{"fediffuselighting", "feDiffuseLighting"},
		{"fedisplacementmap", "feDisplacementMap"},
		{"fedistantlight", "feDistantLight"},
		{"feflood", "feFlood"},
		{"fefunca", "feFuncA"},
		{"fefuncb", "feFuncB"},
		{"fefuncg", "feFuncG"},
		{"fefuncr", "feFuncR"},
		{"fegaussianblur", "feGaussianBlur"},
		{"feimage", "feImage"},
		{"femerge", "feMerge"},
		{"femergenode", "feMergeNode"},
		{"femorphology", "feMorphology"},
		{"feoffset", "feOffset"},
		{"fepointlight", "fePointLight"},
		{"fespecularlighting", "feSpecularLighting"},
		{"fespotlight", "feSpotLight"},
		{"fetile", "feTile"},
		{"feturbulence", "feTurbulence"},
		{"foreignobject", "foreignObject"},
		{"glyphref", "glyphRef"},
		{"lineargradient", "linearGradient"},
		{"radialgradient", "radialGradient"},
		{"textpath", "textPath"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := SVGTagNameAdjustments[tt.input]
			if !ok {
				t.Errorf("SVGTagNameAdjustments[%q] not found", tt.input)
				return
			}
			if got != tt.expected {
				t.Errorf("SVGTagNameAdjustments[%q] = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSVGAttributeAdjustments(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"attributename", "attributeName"},
		{"attributetype", "attributeType"},
		{"basefrequency", "baseFrequency"},
		{"baseprofile", "baseProfile"},
		{"calcmode", "calcMode"},
		{"clippathunits", "clipPathUnits"},
		{"diffuseconstant", "diffuseConstant"},
		{"edgemode", "edgeMode"},
		{"filterunits", "filterUnits"},
		{"glyphref", "glyphRef"},
		{"gradienttransform", "gradientTransform"},
		{"gradientunits", "gradientUnits"},
		{"kernelmatrix", "kernelMatrix"},
		{"kernelunitlength", "kernelUnitLength"},
		{"keypoints", "keyPoints"},
		{"keysplines", "keySplines"},
		{"keytimes", "keyTimes"},
		{"lengthadjust", "lengthAdjust"},
		{"limitingconeangle", "limitingConeAngle"},
		{"markerheight", "markerHeight"},
		{"markerunits", "markerUnits"},
		{"markerwidth", "markerWidth"},
		{"maskcontentunits", "maskContentUnits"},
		{"maskunits", "maskUnits"},
		{"numoctaves", "numOctaves"},
		{"pathlength", "pathLength"},
		{"patterncontentunits", "patternContentUnits"},
		{"patterntransform", "patternTransform"},
		{"patternunits", "patternUnits"},
		{"pointsatx", "pointsAtX"},
		{"pointsaty", "pointsAtY"},
		{"pointsatz", "pointsAtZ"},
		{"preservealpha", "preserveAlpha"},
		{"preserveaspectratio", "preserveAspectRatio"},
		{"primitiveunits", "primitiveUnits"},
		{"refx", "refX"},
		{"refy", "refY"},
		{"repeatcount", "repeatCount"},
		{"repeatdur", "repeatDur"},
		{"requiredextensions", "requiredExtensions"},
		{"requiredfeatures", "requiredFeatures"},
		{"specularconstant", "specularConstant"},
		{"specularexponent", "specularExponent"},
		{"spreadmethod", "spreadMethod"},
		{"startoffset", "startOffset"},
		{"stddeviation", "stdDeviation"},
		{"stitchtiles", "stitchTiles"},
		{"surfacescale", "surfaceScale"},
		{"systemlanguage", "systemLanguage"},
		{"tablevalues", "tableValues"},
		{"targetx", "targetX"},
		{"targety", "targetY"},
		{"textlength", "textLength"},
		{"viewbox", "viewBox"},
		{"viewtarget", "viewTarget"},
		{"xchannelselector", "xChannelSelector"},
		{"ychannelselector", "yChannelSelector"},
		{"zoomandpan", "zoomAndPan"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := SVGAttributeAdjustments[tt.input]
			if !ok {
				t.Errorf("SVGAttributeAdjustments[%q] not found", tt.input)
				return
			}
			if got != tt.expected {
				t.Errorf("SVGAttributeAdjustments[%q] = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMathMLAttributeAdjustments(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"definitionurl", "definitionURL"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := MathMLAttributeAdjustments[tt.input]
			if !ok {
				t.Errorf("MathMLAttributeAdjustments[%q] not found", tt.input)
				return
			}
			if got != tt.expected {
				t.Errorf("MathMLAttributeAdjustments[%q] = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestForeignAttributeAdjustments(t *testing.T) {
	tests := []struct {
		input    string
		expected ForeignAttribute
	}{
		{
			input: "xlink:actuate",
			expected: ForeignAttribute{
				Prefix:       "xlink",
				LocalName:    "actuate",
				NamespaceURL: "http://www.w3.org/1999/xlink",
			},
		},
		{
			input: "xlink:href",
			expected: ForeignAttribute{
				Prefix:       "xlink",
				LocalName:    "href",
				NamespaceURL: "http://www.w3.org/1999/xlink",
			},
		},
		{
			input: "xml:lang",
			expected: ForeignAttribute{
				Prefix:       "xml",
				LocalName:    "lang",
				NamespaceURL: "http://www.w3.org/XML/1998/namespace",
			},
		},
		{
			input: "xml:space",
			expected: ForeignAttribute{
				Prefix:       "xml",
				LocalName:    "space",
				NamespaceURL: "http://www.w3.org/XML/1998/namespace",
			},
		},
		{
			input: "xmlns",
			expected: ForeignAttribute{
				Prefix:       "",
				LocalName:    "xmlns",
				NamespaceURL: "http://www.w3.org/2000/xmlns/",
			},
		},
		{
			input: "xmlns:xlink",
			expected: ForeignAttribute{
				Prefix:       "xmlns",
				LocalName:    "xlink",
				NamespaceURL: "http://www.w3.org/2000/xmlns/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ForeignAttributeAdjustments[tt.input]
			if !ok {
				t.Errorf("ForeignAttributeAdjustments[%q] not found", tt.input)
				return
			}
			if got.Prefix != tt.expected.Prefix {
				t.Errorf("ForeignAttributeAdjustments[%q].Prefix = %q, want %q",
					tt.input, got.Prefix, tt.expected.Prefix)
			}
			if got.LocalName != tt.expected.LocalName {
				t.Errorf("ForeignAttributeAdjustments[%q].LocalName = %q, want %q",
					tt.input, got.LocalName, tt.expected.LocalName)
			}
			if got.NamespaceURL != tt.expected.NamespaceURL {
				t.Errorf("ForeignAttributeAdjustments[%q].NamespaceURL = %q, want %q",
					tt.input, got.NamespaceURL, tt.expected.NamespaceURL)
			}
		})
	}
}

func TestNamespaceConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"HTML", NamespaceHTML, "http://www.w3.org/1999/xhtml"},
		{"SVG", NamespaceSVG, "http://www.w3.org/2000/svg"},
		{"MathML", NamespaceMathML, "http://www.w3.org/1998/Math/MathML"},
		{"XLink", NamespaceXLink, "http://www.w3.org/1999/xlink"},
		{"XML", NamespaceXML, "http://www.w3.org/XML/1998/namespace"},
		{"XMLNS", NamespaceXMLNS, "http://www.w3.org/2000/xmlns/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Namespace%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestHTMLIntegrationPoints(t *testing.T) {
	tests := []struct {
		namespace string
		localName string
		expected  bool
	}{
		{NamespaceMathML, "annotation-xml", true},
		{NamespaceSVG, "foreignObject", true},
		{NamespaceSVG, "desc", true},
		{NamespaceSVG, "title", true},
		{NamespaceSVG, "path", false},
		{NamespaceHTML, "div", false},
	}

	for _, tt := range tests {
		t.Run(tt.namespace+":"+tt.localName, func(t *testing.T) {
			ip := IntegrationPoint{Namespace: tt.namespace, LocalName: tt.localName}
			got := HTMLIntegrationPoints[ip]
			if got != tt.expected {
				t.Errorf("HTMLIntegrationPoints[{%q, %q}] = %v, want %v",
					tt.namespace, tt.localName, got, tt.expected)
			}
		})
	}
}

func TestMathMLTextIntegrationPoints(t *testing.T) {
	tests := []struct {
		namespace string
		localName string
		expected  bool
	}{
		{NamespaceMathML, "mi", true},
		{NamespaceMathML, "mo", true},
		{NamespaceMathML, "mn", true},
		{NamespaceMathML, "ms", true},
		{NamespaceMathML, "mtext", true},
		{NamespaceMathML, "math", false},
		{NamespaceHTML, "div", false},
	}

	for _, tt := range tests {
		t.Run(tt.namespace+":"+tt.localName, func(t *testing.T) {
			ip := IntegrationPoint{Namespace: tt.namespace, LocalName: tt.localName}
			got := MathMLTextIntegrationPoints[ip]
			if got != tt.expected {
				t.Errorf("MathMLTextIntegrationPoints[{%q, %q}] = %v, want %v",
					tt.namespace, tt.localName, got, tt.expected)
			}
		})
	}
}

func TestForeignBreakoutElements(t *testing.T) {
	breakoutElements := []string{
		"b", "big", "blockquote", "body", "br", "center", "code", "dd", "div",
		"dl", "dt", "em", "embed", "h1", "h2", "h3", "h4", "h5", "h6", "head",
		"hr", "i", "img", "li", "listing", "menu", "meta", "nobr", "ol", "p",
		"pre", "ruby", "s", "small", "span", "strong", "strike", "sub", "sup",
		"table", "tt", "u", "ul", "var",
	}

	for _, elem := range breakoutElements {
		t.Run(elem, func(t *testing.T) {
			if !ForeignBreakoutElements[elem] {
				t.Errorf("ForeignBreakoutElements[%q] = false, want true", elem)
			}
		})
	}

	// Test that non-breakout elements are not in the map
	nonBreakout := []string{"svg", "math", "path", "circle", "rect"}
	for _, elem := range nonBreakout {
		t.Run("not-"+elem, func(t *testing.T) {
			if ForeignBreakoutElements[elem] {
				t.Errorf("ForeignBreakoutElements[%q] = true, want false", elem)
			}
		})
	}
}

func TestSVGTagNameAdjustmentsCount(t *testing.T) {
	// Ensure we have all 36 SVG tag name adjustments from the spec
	expected := 36
	got := len(SVGTagNameAdjustments)
	if got != expected {
		t.Errorf("SVGTagNameAdjustments has %d entries, want %d", got, expected)
	}
}

func TestSVGAttributeAdjustmentsCount(t *testing.T) {
	// Ensure we have all 58 SVG attribute adjustments from the spec
	expected := 58
	got := len(SVGAttributeAdjustments)
	if got != expected {
		t.Errorf("SVGAttributeAdjustments has %d entries, want %d", got, expected)
	}
}

func TestForeignAttributeAdjustmentsCount(t *testing.T) {
	// Ensure we have all 11 foreign attribute adjustments from the spec
	expected := 11
	got := len(ForeignAttributeAdjustments)
	if got != expected {
		t.Errorf("ForeignAttributeAdjustments has %d entries, want %d", got, expected)
	}
}

func TestForeignBreakoutElementsCount(t *testing.T) {
	// Ensure we have all 44 foreign breakout elements from the spec
	expected := 44
	got := len(ForeignBreakoutElements)
	if got != expected {
		t.Errorf("ForeignBreakoutElements has %d entries, want %d", got, expected)
	}
}
