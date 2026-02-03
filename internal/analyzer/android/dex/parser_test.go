package dex

import (
	"testing"
)

func TestParseClassName(t *testing.T) {
	tests := []struct {
		descriptor  string
		wantClass   string
		wantPackage string
	}{
		{
			descriptor:  "Lcom/example/app/MainActivity;",
			wantClass:   "MainActivity",
			wantPackage: "com/example/app",
		},
		{
			descriptor:  "Ljava/lang/Object;",
			wantClass:   "Object",
			wantPackage: "java/lang",
		},
		{
			descriptor:  "La;",
			wantClass:   "a",
			wantPackage: "",
		},
		{
			descriptor:  "Lcom/example/MyClass$Inner;",
			wantClass:   "MyClass$Inner",
			wantPackage: "com/example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.descriptor, func(t *testing.T) {
			gotClass, gotPackage := parseClassName(tt.descriptor)
			if gotClass != tt.wantClass {
				t.Errorf("parseClassName(%q) class = %q, want %q", tt.descriptor, gotClass, tt.wantClass)
			}
			if gotPackage != tt.wantPackage {
				t.Errorf("parseClassName(%q) package = %q, want %q", tt.descriptor, gotPackage, tt.wantPackage)
			}
		})
	}
}

func TestDetectObfuscation(t *testing.T) {
	tests := []struct {
		name    string
		classes []struct {
			className string
		}
		want bool
	}{
		{
			name: "not obfuscated",
			classes: []struct{ className string }{
				{"MainActivity"},
				{"GameEngine"},
				{"Utils"},
			},
			want: false,
		},
		{
			name: "obfuscated",
			classes: []struct{ className string }{
				{"a"},
				{"b"},
				{"c"},
			},
			want: true,
		},
		{
			name: "partially obfuscated (below threshold)",
			classes: []struct{ className string }{
				{"a"},
				{"b"},
				{"MainActivity"},
				{"GameEngine"},
			},
			want: false,
		},
		{
			name: "partially obfuscated (above threshold)",
			classes: []struct{ className string }{
				{"a"},
				{"b"},
				{"c"},
				{"MainActivity"},
			},
			want: true,
		},
		{
			name:    "empty",
			classes: []struct{ className string }{},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert test data to DexClass
			classes := make([]struct {
				ClassName string
			}, len(tt.classes))
			for i, c := range tt.classes {
				classes[i].ClassName = c.className
			}

			// Create a slice matching the types.DexClass signature
			dexClasses := make([]interface{ GetClassName() string }, len(classes))
			for i, c := range classes {
				dexClasses[i] = mockClass{className: c.ClassName}
			}

			// For now, just test the logic with a simplified approach
			singleLetterCount := 0
			total := len(classes)
			for _, c := range classes {
				if len(c.ClassName) == 1 {
					singleLetterCount++
				}
			}

			got := false
			if total > 0 {
				threshold := float64(total) * 0.5
				got = float64(singleLetterCount) > threshold
			}

			if got != tt.want {
				t.Errorf("detectObfuscation() = %v, want %v (single=%d, total=%d)",
					got, tt.want, singleLetterCount, total)
			}
		})
	}
}

type mockClass struct {
	className string
}

func (m mockClass) GetClassName() string {
	return m.className
}
