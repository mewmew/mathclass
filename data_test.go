package mathclass

import (
	"testing"
	"unicode"
)

func TestMathClass(t *testing.T) {
	golden := []struct {
		input rune
		class *unicode.RangeTable
	}{
		{input: '0', class: Normal},
		{input: 'a', class: Alphabetic},
		{input: '𝔸', class: Alphabetic},
		{input: '+', class: Vary},
		{input: '×', class: Binary},
		{input: '(', class: Opening},
		{input: ',', class: Punctuation},
		{input: '|', class: Fence},
		{input: '😃', class: nil},
	}
	classes := []*unicode.RangeTable{
		Normal,
		Alphabetic,
		Binary,
		Closing,
		Diacritic,
		Fence,
		Glyph_Part,
		Large,
		Opening,
		Punctuation,
		Relation,
		Space,
		Unary,
		Vary,
		Special,
	}
	for _, g := range golden {
		for _, class := range classes {
			if unicode.In(g.input, class) {
				if g.class == nil {
					t.Errorf("input %q should not be identified by class %q", g.input, class)
					continue
				}
				if g.class != class {
					t.Errorf("class mismatch for %q; expected %v, got %v", g.input, g.class, class)
				}
			}
		}
	}
}
