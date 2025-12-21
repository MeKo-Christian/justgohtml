package tokenizer

import (
	"testing"
)

const (
	// Pure ASCII HTML document for benchmarking
	asciiHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ASCII Test Document</title>
    <link rel="stylesheet" href="styles.css">
    <script src="app.js"></script>
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
            <h1>Hello World</h1>
            <p class="intro">This is a test document with pure ASCII content.</p>
            <div class="content">
                <p>Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>
                <p>Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p>
            </div>
        </article>
    </main>
    <footer>
        <p>&copy; 2024 Test Site. All rights reserved.</p>
    </footer>
</body>
</html>`

	// Unicode HTML document for comparison
	//nolint:gosmopolitan
	unicodeHTML = `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <title>Unicode テスト</title>
</head>
<body>
    <div class="コンテナ">
        <h1>こんにちは世界</h1>
        <p>これはUnicodeコンテンツです。</p>
        <p>日本語、中文、한글、العربية</p>
    </div>
</body>
</html>`
)

// BenchmarkTokenizer_ASCII_FastPath benchmarks tokenization with ASCII fast path enabled
func BenchmarkTokenizer_ASCII_FastPath(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		tok := New(asciiHTML)
		for {
			token := tok.Next()
			if token.Type == EOF {
				break
			}
		}
	}
}

// BenchmarkTokenizer_ASCII_ForceRuneMode benchmarks ASCII content with rune mode forced
func BenchmarkTokenizer_ASCII_ForceRuneMode(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		tok := New(asciiHTML)
		tok.ForceRuneMode() // Disable ASCII fast path for comparison
		for {
			token := tok.Next()
			if token.Type == EOF {
				break
			}
		}
	}
}

// BenchmarkTokenizer_Unicode benchmarks tokenization with Unicode content
func BenchmarkTokenizer_Unicode(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		tok := New(unicodeHTML)
		for {
			token := tok.Next()
			if token.Type == EOF {
				break
			}
		}
	}
}

// BenchmarkGetChar_ASCII benchmarks the ASCII-optimized getChar
func BenchmarkGetChar_ASCII(b *testing.B) {
	tok := New(asciiHTML)
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		tok.reset(asciiHTML)
		for {
			_, ok := tok.getChar()
			if !ok {
				break
			}
		}
	}
}

// BenchmarkGetChar_Rune benchmarks the rune-based getChar
func BenchmarkGetChar_Rune(b *testing.B) {
	tok := New(asciiHTML)
	tok.ForceRuneMode()
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		tok.reset(asciiHTML)
		tok.ForceRuneMode()
		for {
			_, ok := tok.getChar()
			if !ok {
				break
			}
		}
	}
}
