package serialize

import (
	"testing"

	"github.com/MeKo-Christian/JustGoHTML"
	"github.com/MeKo-Christian/JustGoHTML/dom"
)

// Test HTML samples for serialization benchmarks
const (
	simpleHTML = `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<div id="main">
<p class="intro">Hello, World!</p>
<ul>
<li>Item 1</li>
<li>Item 2</li>
<li>Item 3</li>
</ul>
</div>
</body>
</html>`

	mediumHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Blog Post - Example Site</title>
    <link rel="stylesheet" href="styles.css">
</head>
<body>
    <header>
        <nav>
            <ul>
                <li><a href="/">Home</a></li>
                <li><a href="/about">About</a></li>
                <li><a href="/blog">Blog</a></li>
                <li><a href="/contact">Contact</a></li>
            </ul>
        </nav>
    </header>
    <main>
        <article>
            <h1>Understanding HTML5 Parsing</h1>
            <p class="meta">Published on <time datetime="2025-01-15">January 15, 2025</time></p>
            <section>
                <h2>Introduction</h2>
                <p>The HTML5 specification defines how browsers should parse HTML documents.</p>
                <ul>
                    <li>Error recovery rules</li>
                    <li>Tree construction algorithms</li>
                    <li>Tokenization state machines</li>
                </ul>
            </section>
        </article>
    </main>
    <footer>
        <p>&copy; 2025 Example Site. All rights reserved.</p>
    </footer>
</body>
</html>`

	complexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta property="og:title" content="Complex Page">
    <title>Complex HTML Page</title>
    <style>
        body { font-family: Arial, sans-serif; }
        .container { max-width: 1200px; margin: 0 auto; }
    </style>
    <script>
        console.log('Page loaded');
        var data = { key: 'value' };
    </script>
</head>
<body>
    <div class="container">
        <header>
            <nav aria-label="Main navigation">
                <ul class="nav-list">
                    <li><a href="/" aria-current="page">Home</a></li>
                    <li><a href="/products">Products</a></li>
                    <li><a href="/services">Services</a></li>
                </ul>
            </nav>
        </header>
        <main>
            <section id="hero">
                <h1>Welcome to Our Website</h1>
                <p class="lead">We provide excellent services</p>
            </section>
            <section id="features">
                <h2>Our Features</h2>
                <div class="feature-grid">
                    <div class="feature" data-feature-id="1">
                        <h3>Fast Performance</h3>
                        <p>Optimized for speed</p>
                    </div>
                    <div class="feature" data-feature-id="2">
                        <h3>Reliable</h3>
                        <p>99.9% uptime guaranteed</p>
                    </div>
                </div>
            </section>
            <section id="contact">
                <h2>Contact Us</h2>
                <form action="/submit" method="post">
                    <div class="form-group">
                        <label for="name">Name:</label>
                        <input type="text" id="name" name="name" required>
                    </div>
                    <div class="form-group">
                        <label for="email">Email:</label>
                        <input type="email" id="email" name="email" required>
                    </div>
                    <button type="submit">Send</button>
                </form>
            </section>
        </main>
        <footer>
            <p class="copyright">&copy; 2025 Example Corp.</p>
        </footer>
    </div>
</body>
</html>`
)

// BenchmarkToHTML_Simple benchmarks serialization of simple HTML
func BenchmarkToHTML_Simple(b *testing.B) {
	doc, err := JustGoHTML.Parse(simpleHTML)
	if err != nil {
		b.Fatal(err)
	}

	opts := DefaultOptions()
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		html := ToHTML(doc, opts)
		_ = html
	}
}

// BenchmarkToHTML_Medium benchmarks serialization of medium complexity HTML
func BenchmarkToHTML_Medium(b *testing.B) {
	doc, err := JustGoHTML.Parse(mediumHTML)
	if err != nil {
		b.Fatal(err)
	}

	opts := DefaultOptions()
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		html := ToHTML(doc, opts)
		_ = html
	}
}

// BenchmarkToHTML_Complex benchmarks serialization of complex HTML
func BenchmarkToHTML_Complex(b *testing.B) {
	doc, err := JustGoHTML.Parse(complexHTML)
	if err != nil {
		b.Fatal(err)
	}

	opts := DefaultOptions()
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		html := ToHTML(doc, opts)
		_ = html
	}
}

// BenchmarkToHTML_Pretty benchmarks pretty-printed serialization
func BenchmarkToHTML_Pretty(b *testing.B) {
	doc, err := JustGoHTML.Parse(complexHTML)
	if err != nil {
		b.Fatal(err)
	}

	opts := Options{Pretty: true, IndentSize: 2}
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		html := ToHTML(doc, opts)
		_ = html
	}
}

// BenchmarkToHTML_Element benchmarks serialization of a single element
func BenchmarkToHTML_Element(b *testing.B) {
	doc, err := JustGoHTML.Parse(complexHTML)
	if err != nil {
		b.Fatal(err)
	}

	// Find a complex element
	elements, err := doc.Query("section#features")
	if err != nil || len(elements) == 0 {
		b.Fatal("Could not find test element")
	}

	elem := elements[0]
	opts := DefaultOptions()
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		html := ToHTML(elem, opts)
		_ = html
	}
}

// BenchmarkToHTML_DeepNesting benchmarks serialization of deeply nested structures
func BenchmarkToHTML_DeepNesting(b *testing.B) {
	// Create a deeply nested structure
	doc := dom.NewDocument()
	root := dom.NewElement("div")
	doc.AppendChild(root)

	current := root
	for range 20 {
		child := dom.NewElement("div")
		child.SetAttr("class", "nested")
		text := dom.NewText("Content")
		child.AppendChild(text)
		current.AppendChild(child)
		current = child
	}

	opts := DefaultOptions()
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		html := ToHTML(doc, opts)
		_ = html
	}
}

// BenchmarkToHTML_ManyAttributes benchmarks serialization of elements with many attributes
func BenchmarkToHTML_ManyAttributes(b *testing.B) {
	doc := dom.NewDocument()
	elem := dom.NewElement("div")

	// Add many attributes
	for i := range 50 {
		elem.SetAttr("data-attr-"+string(rune('0'+(i%10))), "value")
	}
	elem.AppendChild(dom.NewText("Content"))
	doc.AppendChild(elem)

	opts := DefaultOptions()
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		html := ToHTML(doc, opts)
		_ = html
	}
}

// BenchmarkToHTML_LargeText benchmarks serialization with large text nodes
func BenchmarkToHTML_LargeText(b *testing.B) {
	doc := dom.NewDocument()
	elem := dom.NewElement("p")

	// Create large text content
	largeText := make([]byte, 10000)
	for i := range largeText {
		largeText[i] = 'a' + byte(i%26)
	}
	elem.AppendChild(dom.NewText(string(largeText)))
	doc.AppendChild(elem)

	opts := DefaultOptions()
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		html := ToHTML(doc, opts)
		_ = html
	}
}

// BenchmarkToHTML_SpecialChars benchmarks serialization with many special characters
func BenchmarkToHTML_SpecialChars(b *testing.B) {
	doc := dom.NewDocument()
	elem := dom.NewElement("p")
	elem.SetAttr("title", `Special "quotes" & <tags>`)
	elem.AppendChild(dom.NewText(`Text with <special> & "characters" that need escaping`))
	doc.AppendChild(elem)

	opts := DefaultOptions()
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		html := ToHTML(doc, opts)
		_ = html
	}
}

// BenchmarkToHTML_ManyChildren benchmarks serialization of elements with many children
func BenchmarkToHTML_ManyChildren(b *testing.B) {
	doc := dom.NewDocument()
	ul := dom.NewElement("ul")

	// Create many list items
	for i := range 100 {
		li := dom.NewElement("li")
		li.AppendChild(dom.NewText("Item " + string(rune('0'+(i%10)))))
		ul.AppendChild(li)
	}
	doc.AppendChild(ul)

	opts := DefaultOptions()
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		html := ToHTML(doc, opts)
		_ = html
	}
}

// BenchmarkToHTML_Script benchmarks serialization of script elements
func BenchmarkToHTML_Script(b *testing.B) {
	doc := dom.NewDocument()
	script := dom.NewElement("script")
	script.AppendChild(dom.NewText(`
		function example() {
			var x = '<div>';
			console.log("test");
			return x + '</div>';
		}
	`))
	doc.AppendChild(script)

	opts := DefaultOptions()
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		html := ToHTML(doc, opts)
		_ = html
	}
}

// BenchmarkToHTML_Style benchmarks serialization of style elements
func BenchmarkToHTML_Style(b *testing.B) {
	doc := dom.NewDocument()
	style := dom.NewElement("style")
	style.AppendChild(dom.NewText(`
		body { margin: 0; padding: 0; }
		.container { max-width: 1200px; }
		.feature > h3 { color: blue; }
	`))
	doc.AppendChild(style)

	opts := DefaultOptions()
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		html := ToHTML(doc, opts)
		_ = html
	}
}

// BenchmarkToHTML_Parallel benchmarks parallel serialization
func BenchmarkToHTML_Parallel(b *testing.B) {
	doc, err := JustGoHTML.Parse(complexHTML)
	if err != nil {
		b.Fatal(err)
	}

	opts := DefaultOptions()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			html := ToHTML(doc, opts)
			_ = html
		}
	})
}
