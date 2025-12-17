package constants

import (
	"testing"
)

// TestNamedEntitiesCount verifies we have all 2,125 HTML5 entities
func TestNamedEntitiesCount(t *testing.T) {
	expected := 2125
	actual := len(NamedEntities)
	if actual != expected {
		t.Errorf("Expected %d entities, got %d", expected, actual)
	}
}

// TestNamedEntitiesBasic tests common named entities
func TestNamedEntitiesBasic(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"amp", "&"},
		{"lt", "<"},
		{"gt", ">"},
		{"quot", "\""},
		{"nbsp", "\u00A0"},
		{"copy", "©"},
		{"reg", "®"},
		{"AElig", "Æ"},
		{"aelig", "æ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, ok := NamedEntities[tt.name]
			if !ok {
				t.Errorf("Entity %q not found", tt.name)
				return
			}
			if actual != tt.expected {
				t.Errorf("Entity %q: expected %q, got %q", tt.name, tt.expected, actual)
			}
		})
	}
}

// TestNamedEntitiesMultiChar tests entities that decode to multiple characters
func TestNamedEntitiesMultiChar(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"NotEqualTilde", "≂̸"},
		{"acE", "∾̳"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, ok := NamedEntities[tt.name]
			if !ok {
				t.Errorf("Entity %q not found", tt.name)
				return
			}
			if actual != tt.expected {
				t.Errorf("Entity %q: expected %q, got %q", tt.name, tt.expected, actual)
			}
		})
	}
}

// TestNamedEntitiesCaseSensitive verifies entity names are case-sensitive
func TestNamedEntitiesCaseSensitive(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"Alpha", "Α"}, // Greek capital alpha
		{"alpha", "α"}, // Greek lowercase alpha
		{"COPY", "©"},
		{"copy", "©"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, ok := NamedEntities[tt.name]
			if !ok {
				t.Errorf("Entity %q not found", tt.name)
				return
			}
			if actual != tt.expected {
				t.Errorf("Entity %q: expected %q, got %q", tt.name, tt.expected, actual)
			}
		})
	}
}

// TestLegacyEntitiesCount verifies we have all legacy entities
func TestLegacyEntitiesCount(t *testing.T) {
	expected := 106 // As defined in Python reference
	actual := len(LegacyEntities)
	if actual != expected {
		t.Errorf("Expected %d legacy entities, got %d", expected, actual)
	}
}

// TestLegacyEntitiesBasic tests that common legacy entities are present
func TestLegacyEntitiesBasic(t *testing.T) {
	tests := []string{
		"amp", "lt", "gt", "quot", "nbsp",
		"copy", "reg", "AElig", "aacute",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			if !LegacyEntities[name] {
				t.Errorf("Legacy entity %q not found", name)
			}
		})
	}
}

// TestLegacyEntitiesAreInNamedEntities verifies all legacy entities exist in named entities
func TestLegacyEntitiesAreInNamedEntities(t *testing.T) {
	for name := range LegacyEntities {
		if _, ok := NamedEntities[name]; !ok {
			t.Errorf("Legacy entity %q not found in NamedEntities", name)
		}
	}
}

// TestModernEntitiesNotInLegacy verifies modern entities are not in legacy set
func TestModernEntitiesNotInLegacy(t *testing.T) {
	// Modern entities that require semicolons
	modern := []string{
		"lang",  // ⟨
		"rang",  // ⟩
		"notin", // ∉
		"prod",  // ∏
	}

	for _, name := range modern {
		t.Run(name, func(t *testing.T) {
			// Should exist in NamedEntities
			if _, ok := NamedEntities[name]; !ok {
				t.Errorf("Modern entity %q not found in NamedEntities", name)
			}
			// Should NOT be in LegacyEntities
			if LegacyEntities[name] {
				t.Errorf("Modern entity %q incorrectly in LegacyEntities", name)
			}
		})
	}
}

// TestNumericReplacementsCount verifies we have all replacements
func TestNumericReplacementsCount(t *testing.T) {
	expected := 28 // As defined in Python reference
	actual := len(NumericReplacements)
	if actual != expected {
		t.Errorf("Expected %d numeric replacements, got %d", expected, actual)
	}
}

// TestNumericReplacementsBasic tests common numeric replacements
func TestNumericReplacementsBasic(t *testing.T) {
	tests := []struct {
		code     int
		expected rune
	}{
		{0x00, '\uFFFD'}, // NULL -> REPLACEMENT CHARACTER
		{0x80, '\u20AC'}, // EURO SIGN
		{0x82, '\u201A'}, // SINGLE LOW-9 QUOTATION MARK
		{0x91, '\u2018'}, // LEFT SINGLE QUOTATION MARK
		{0x92, '\u2019'}, // RIGHT SINGLE QUOTATION MARK
		{0x99, '\u2122'}, // TRADE MARK SIGN
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			actual, ok := NumericReplacements[tt.code]
			if !ok {
				t.Errorf("Numeric replacement for 0x%02X not found", tt.code)
				return
			}
			if actual != tt.expected {
				t.Errorf("Numeric replacement for 0x%02X: expected %q, got %q", tt.code, tt.expected, actual)
			}
		})
	}
}

// TestNumericReplacementsCompleteness verifies all expected codes are present
func TestNumericReplacementsCompleteness(t *testing.T) {
	expectedCodes := []int{
		0x00, 0x80, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89,
		0x8A, 0x8B, 0x8C, 0x8E, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96,
		0x97, 0x98, 0x99, 0x9A, 0x9B, 0x9C, 0x9E, 0x9F,
	}

	for _, code := range expectedCodes {
		if _, ok := NumericReplacements[code]; !ok {
			t.Errorf("Expected numeric replacement for 0x%02X not found", code)
		}
	}
}

// TestSpecificNamedEntities tests specific entities from html5lib tests
func TestSpecificNamedEntities(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"not", "¬"},      // Legacy entity
		{"lang", "⟨"},     // Modern entity requiring semicolon
		{"rang", "⟩"},     // Modern entity requiring semicolon
		{"notin", "∉"},    // Modern entity requiring semicolon
		{"prod", "∏"},     // Modern entity requiring semicolon
		{"NewLine", "\n"}, // Contains newline character
		{"Tab", "\t"},     // Contains tab character
		{"ZeroWidthSpace", "\u200B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, ok := NamedEntities[tt.name]
			if !ok {
				t.Errorf("Entity %q not found", tt.name)
				return
			}
			if actual != tt.expected {
				t.Errorf("Entity %q: expected %+q, got %+q", tt.name, tt.expected, actual)
			}
		})
	}
}

// TestNonExistentEntities verifies that certain entity names do NOT exist
func TestNonExistentEntities(t *testing.T) {
	nonExistent := []string{
		"noti", // Similar to "not" but not a valid entity
	}

	for _, name := range nonExistent {
		t.Run(name, func(t *testing.T) {
			if _, ok := NamedEntities[name]; ok {
				t.Errorf("Entity %q should not exist but was found", name)
			}
		})
	}
}

// Benchmarks for entity lookup performance

// BenchmarkNamedEntityLookupCommon benchmarks lookup of common entities
func BenchmarkNamedEntityLookupCommon(b *testing.B) {
	commonEntities := []string{"amp", "lt", "gt", "quot", "nbsp"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := commonEntities[i%len(commonEntities)]
		_, _ = NamedEntities[name]
	}
}

// BenchmarkNamedEntityLookupUncommon benchmarks lookup of uncommon entities
func BenchmarkNamedEntityLookupUncommon(b *testing.B) {
	uncommonEntities := []string{"NotEqualTilde", "acE", "lang", "rang", "notin"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := uncommonEntities[i%len(uncommonEntities)]
		_, _ = NamedEntities[name]
	}
}

// BenchmarkNamedEntityLookupMissing benchmarks lookup of non-existent entities
func BenchmarkNamedEntityLookupMissing(b *testing.B) {
	missingEntities := []string{"notanentity", "invalid", "xyz", "test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := missingEntities[i%len(missingEntities)]
		_, _ = NamedEntities[name]
	}
}

// BenchmarkLegacyEntityLookup benchmarks lookup in the legacy entities map
func BenchmarkLegacyEntityLookup(b *testing.B) {
	legacyNames := []string{"amp", "lt", "gt", "quot", "nbsp", "copy", "reg"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := legacyNames[i%len(legacyNames)]
		_ = LegacyEntities[name]
	}
}

// BenchmarkNumericReplacementLookup benchmarks numeric replacement lookup
func BenchmarkNumericReplacementLookup(b *testing.B) {
	codes := []int{0x00, 0x80, 0x82, 0x91, 0x92, 0x99}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		code := codes[i%len(codes)]
		_, _ = NumericReplacements[code]
	}
}

// BenchmarkNamedEntityLookupAll benchmarks sequential lookup of all entities
func BenchmarkNamedEntityLookupAll(b *testing.B) {
	// Create a slice of all entity names for sequential access
	names := make([]string, 0, len(NamedEntities))
	for name := range NamedEntities {
		names = append(names, name)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := names[i%len(names)]
		_, _ = NamedEntities[name]
	}
}

// BenchmarkNamedEntityLookupByLength benchmarks lookup by entity name length
func BenchmarkNamedEntityLookupByLength(b *testing.B) {
	tests := []struct {
		name     string
		entities []string
	}{
		{"Short", []string{"lt", "gt", "pi", "mu", "nu"}},
		{"Medium", []string{"copy", "nbsp", "lang", "rang"}},
		{"Long", []string{"NotEqualTilde", "DoubleLeftTee", "TripleDot"}},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				name := tt.entities[i%len(tt.entities)]
				_, _ = NamedEntities[name]
			}
		})
	}
}
