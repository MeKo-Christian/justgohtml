package constants

import (
	"testing"
	"unsafe"
)

func TestInternTagName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantSame bool // Should return the same pointer as in CommonTagNames
	}{
		{"common tag div", "div", true},
		{"common tag span", "span", true},
		{"common tag p", "p", true},
		{"common tag html", "html", true},
		{"uncommon tag custom-element", "custom-element", false},
		{"uncommon tag mywidget", "mywidget", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InternTagName(tt.input)

			// Check that the value is correct
			if got != tt.input {
				t.Errorf("InternTagName(%q) = %q, want %q", tt.input, got, tt.input)
			}

			// Check pointer equality for common tags
			if tt.wantSame {
				expected, ok := CommonTagNames[tt.input]
				if !ok {
					t.Fatalf("Test setup error: %q should be in CommonTagNames", tt.input)
				}
				// Use pointer comparison to verify we're returning the same string instance
				if unsafe.StringData(got) != unsafe.StringData(expected) {
					t.Errorf("InternTagName(%q) did not return interned string", tt.input)
				}
			}
		})
	}
}

func TestInternAttributeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantSame bool // Should return the same pointer as in CommonAttributeNames
	}{
		{"common attr id", "id", true},
		{"common attr class", "class", true},
		{"common attr href", "href", true},
		{"common attr src", "src", true},
		{"uncommon attr data-custom-id", "data-custom-id", false},
		{"uncommon attr ng-model", "ng-model", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InternAttributeName(tt.input)

			// Check that the value is correct
			if got != tt.input {
				t.Errorf("InternAttributeName(%q) = %q, want %q", tt.input, got, tt.input)
			}

			// Check pointer equality for common attributes
			if tt.wantSame {
				expected, ok := CommonAttributeNames[tt.input]
				if !ok {
					t.Fatalf("Test setup error: %q should be in CommonAttributeNames", tt.input)
				}
				// Use pointer comparison to verify we're returning the same string instance
				if unsafe.StringData(got) != unsafe.StringData(expected) {
					t.Errorf("InternAttributeName(%q) did not return interned string", tt.input)
				}
			}
		})
	}
}

func TestCommonTagNamesCoverage(t *testing.T) {
	// Verify all entries in CommonTagNames map to themselves
	for key, value := range CommonTagNames {
		if key != value {
			t.Errorf("CommonTagNames[%q] = %q, want %q", key, value, key)
		}
		// Verify they're the same string instance (interned)
		if unsafe.StringData(key) != unsafe.StringData(value) {
			t.Errorf("CommonTagNames[%q] is not interned (different string instances)", key)
		}
	}
}

func TestCommonAttributeNamesCoverage(t *testing.T) {
	// Verify all entries in CommonAttributeNames map to themselves
	for key, value := range CommonAttributeNames {
		if key != value {
			t.Errorf("CommonAttributeNames[%q] = %q, want %q", key, value, key)
		}
		// Verify they're the same string instance (interned)
		if unsafe.StringData(key) != unsafe.StringData(value) {
			t.Errorf("CommonAttributeNames[%q] is not interned (different string instances)", key)
		}
	}
}

func BenchmarkInternTagName(b *testing.B) {
	b.Run("common tag", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			_ = InternTagName("div")
		}
	})

	b.Run("uncommon tag", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			_ = InternTagName("custom-element")
		}
	})
}

func BenchmarkInternAttributeName(b *testing.B) {
	b.Run("common attr", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			_ = InternAttributeName("class")
		}
	})

	b.Run("uncommon attr", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			_ = InternAttributeName("data-custom-id")
		}
	})
}
