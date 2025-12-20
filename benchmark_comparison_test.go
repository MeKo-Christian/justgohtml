package JustGoHTML

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// Test HTML samples for benchmarking
const (
	// Simple HTML document
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

	// Medium complexity HTML (simulating a blog post)
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
            <p class="meta">Published on <time datetime="2025-01-15">January 15, 2025</time> by <span class="author">John Doe</span></p>
            <p class="intro">HTML5 parsing is a complex topic that involves understanding the WHATWG specification...</p>
            <section>
                <h2>Introduction</h2>
                <p>The HTML5 specification defines how browsers should parse HTML documents. This includes handling malformed HTML, which is surprisingly common.</p>
                <ul>
                    <li>Error recovery rules</li>
                    <li>Tree construction algorithms</li>
                    <li>Tokenization state machines</li>
                </ul>
            </section>
            <section>
                <h2>Key Concepts</h2>
                <p>Several important concepts are central to HTML5 parsing:</p>
                <ol>
                    <li><strong>Tokenization</strong>: Breaking the input into tokens</li>
                    <li><strong>Tree Construction</strong>: Building the DOM tree from tokens</li>
                    <li><strong>Error Handling</strong>: Recovering from malformed markup</li>
                </ol>
            </section>
            <section>
                <h2>Code Example</h2>
                <pre><code class="language-go">
doc, err := JustGoHTML.Parse(html)
if err != nil {
    log.Fatal(err)
}
elements := doc.Query("p.intro")
                </code></pre>
            </section>
        </article>
        <aside>
            <h3>Related Posts</h3>
            <ul>
                <li><a href="/post1">DOM Manipulation in Go</a></li>
                <li><a href="/post2">CSS Selectors Guide</a></li>
                <li><a href="/post3">Web Scraping Best Practices</a></li>
            </ul>
        </aside>
    </main>
    <footer>
        <p>&copy; 2025 Example Site. All rights reserved.</p>
    </footer>
</body>
</html>`

	// Complex HTML with nested structures and special elements
	complexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta property="og:title" content="Complex Page">
    <meta property="og:description" content="A complex HTML page for benchmarking">
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
                    <li><a href="/about">About</a></li>
                </ul>
            </nav>
        </header>
        <main>
            <section id="hero">
                <h1>Welcome to Our Website</h1>
                <p class="lead">We provide excellent services</p>
                <button class="cta-button" data-action="signup">Get Started</button>
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
                    <div class="feature" data-feature-id="3">
                        <h3>Secure</h3>
                        <p>Bank-level encryption</p>
                    </div>
                </div>
            </section>
            <section id="testimonials">
                <h2>What Our Customers Say</h2>
                <blockquote cite="https://example.com/testimonial1">
                    <p>This service changed our business!</p>
                    <footer>— Jane Smith, <cite>ABC Corp</cite></footer>
                </blockquote>
                <blockquote cite="https://example.com/testimonial2">
                    <p>Highly recommended for anyone looking for quality.</p>
                    <footer>— Bob Johnson, <cite>XYZ Inc</cite></footer>
                </blockquote>
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
                    <div class="form-group">
                        <label for="message">Message:</label>
                        <textarea id="message" name="message" rows="5"></textarea>
                    </div>
                    <button type="submit">Send</button>
                </form>
            </section>
        </main>
        <footer>
            <div class="footer-content">
                <div class="footer-section">
                    <h4>About Us</h4>
                    <p>We are dedicated to providing the best service possible.</p>
                </div>
                <div class="footer-section">
                    <h4>Quick Links</h4>
                    <ul>
                        <li><a href="/privacy">Privacy Policy</a></li>
                        <li><a href="/terms">Terms of Service</a></li>
                        <li><a href="/contact">Contact</a></li>
                    </ul>
                </div>
                <div class="footer-section">
                    <h4>Follow Us</h4>
                    <ul class="social-links">
                        <li><a href="https://twitter.com/example">Twitter</a></li>
                        <li><a href="https://facebook.com/example">Facebook</a></li>
                        <li><a href="https://linkedin.com/company/example">LinkedIn</a></li>
                    </ul>
                </div>
            </div>
            <p class="copyright">&copy; 2025 Example Corp. All rights reserved.</p>
        </footer>
    </div>
</body>
</html>`
)

// =============================================================================
// JustGoHTML Benchmarks
// =============================================================================

func BenchmarkJustGoHTML_Parse_Simple(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		doc, err := Parse(simpleHTML)
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

func BenchmarkJustGoHTML_Parse_Medium(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		doc, err := Parse(mediumHTML)
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

func BenchmarkJustGoHTML_Parse_Complex(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		doc, err := Parse(complexHTML)
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

func BenchmarkJustGoHTML_ParseBytes_Simple(b *testing.B) {
	data := []byte(simpleHTML)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		doc, err := ParseBytes(data)
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

func BenchmarkJustGoHTML_ParseBytes_Medium(b *testing.B) {
	data := []byte(mediumHTML)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		doc, err := ParseBytes(data)
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

func BenchmarkJustGoHTML_ParseBytes_Complex(b *testing.B) {
	data := []byte(complexHTML)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		doc, err := ParseBytes(data)
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

func BenchmarkJustGoHTML_Query_Simple(b *testing.B) {
	doc, err := Parse(complexHTML)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		results, err := doc.Query("div.feature")
		if err != nil {
			b.Fatal(err)
		}
		_ = results
	}
}

func BenchmarkJustGoHTML_Query_Complex(b *testing.B) {
	doc, err := Parse(complexHTML)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		results, err := doc.Query("section > h2 + div.feature-grid div[data-feature-id]")
		if err != nil {
			b.Fatal(err)
		}
		_ = results
	}
}

// =============================================================================
// golang.org/x/net/html Benchmarks
// =============================================================================

func BenchmarkNetHTML_Parse_Simple(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		doc, err := html.Parse(strings.NewReader(simpleHTML))
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

func BenchmarkNetHTML_Parse_Medium(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		doc, err := html.Parse(strings.NewReader(mediumHTML))
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

func BenchmarkNetHTML_Parse_Complex(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		doc, err := html.Parse(strings.NewReader(complexHTML))
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

// =============================================================================
// goquery Benchmarks
// =============================================================================

func BenchmarkGoquery_Parse_Simple(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(simpleHTML))
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

func BenchmarkGoquery_Parse_Medium(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(mediumHTML))
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

func BenchmarkGoquery_Parse_Complex(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(complexHTML))
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

func BenchmarkGoquery_Query_Simple(b *testing.B) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(complexHTML))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		selection := doc.Find("div.feature")
		_ = selection
	}
}

func BenchmarkGoquery_Query_Complex(b *testing.B) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(complexHTML))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		selection := doc.Find("section > h2 + div.feature-grid div[data-feature-id]")
		_ = selection
	}
}

// =============================================================================
// Memory Allocation Benchmarks
// =============================================================================

func BenchmarkJustGoHTML_ParseBytes_AllocsPerOp(b *testing.B) {
	data := []byte(complexHTML)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		doc, err := ParseBytes(data)
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

func BenchmarkNetHTML_Parse_AllocsPerOp(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		doc, err := html.Parse(strings.NewReader(complexHTML))
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

func BenchmarkGoquery_Parse_AllocsPerOp(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(complexHTML))
		if err != nil {
			b.Fatal(err)
		}
		_ = doc
	}
}

// =============================================================================
// Parallel Benchmarks
// =============================================================================

func BenchmarkJustGoHTML_Parse_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doc, err := Parse(complexHTML)
			if err != nil {
				b.Fatal(err)
			}
			_ = doc
		}
	})
}

func BenchmarkNetHTML_Parse_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doc, err := html.Parse(strings.NewReader(complexHTML))
			if err != nil {
				b.Fatal(err)
			}
			_ = doc
		}
	})
}

func BenchmarkGoquery_Parse_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(complexHTML))
			if err != nil {
				b.Fatal(err)
			}
			_ = doc
		}
	})
}
