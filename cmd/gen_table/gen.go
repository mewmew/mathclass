package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/pkg/errors"
)

func main() {
	// https://www.unicode.org/Public/math/revision-15/MathClass-15.txt
	if err := parse("MathClass-15.txt"); err != nil {
		log.Fatalf("%+v", err)
	}
	if err := output("data.go"); err != nil {
		log.Fatalf("%+v", err)
	}
}

func output(path string) error {
	out := &bytes.Buffer{}
	t, err := template.ParseFiles("data.go.tmpl")
	if err != nil {
		return errors.WithStack(err)
	}
	data := map[string]any{
		"Normal":      Normal,
		"Alphabetic":  Alphabetic,
		"Binary":      Binary,
		"Closing":     Closing,
		"Diacritic":   Diacritic,
		"Fence":       Fence,
		"Glyph_Part":  Glyph_Part,
		"Large":       Large,
		"Opening":     Opening,
		"Punctuation": Punctuation,
		"Relation":    Relation,
		"Space":       Space,
		"Unary":       Unary,
		"Vary":        Vary,
		"Special":     Special,
	}
	if err := t.Execute(out, data); err != nil {
		return errors.WithStack(err)
	}
	src, err := format.Source(out.Bytes())
	if err != nil {
		return errors.WithStack(err)
	}
	if err := os.WriteFile(path, src, 0o644); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Example:
//
//	002F;B
//	0030..0039;N
func parse(path string) error {
	buf, err := os.ReadFile(path)
	if err != nil {
		return errors.WithStack(err)
	}
	r := bytes.NewReader(buf)
	s := bufio.NewScanner(r)
	for s.Scan() {
		text := s.Text()
		if strings.HasPrefix(text, "#") {
			continue
		}
		if len(text) == 0 {
			continue
		}
		parts := strings.Split(text, ";")
		if len(parts) != 2 {
			return errors.Errorf("expected two semi-colon delimited parts, got %v", text)
		}
		raw_range := parts[0]
		raw_class := parts[1]
		class := getClass(raw_class)
		if err := addRange(class, raw_range); err != nil {
			return errors.WithStack(err)
		}
	}
	if err := s.Err(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

var (
	Normal      = &unicode.RangeTable{} // Normal - includes all digits and symbols requiring only one form
	Alphabetic  = &unicode.RangeTable{} // Alphabetic
	Binary      = &unicode.RangeTable{} // Binary
	Closing     = &unicode.RangeTable{} // Closing - usually paired with opening delimiter
	Diacritic   = &unicode.RangeTable{} // Diacritic
	Fence       = &unicode.RangeTable{} // Fence - unpaired delimiter (often used as opening or closing)
	Glyph_Part  = &unicode.RangeTable{} // Glyph_Part - piece of large operator
	Large       = &unicode.RangeTable{} // Large - n-ary or large operator, often takes limits
	Opening     = &unicode.RangeTable{} // Opening - usually paired with closing delimiter
	Punctuation = &unicode.RangeTable{} // Punctuation
	Relation    = &unicode.RangeTable{} // Relation - includes arrows
	Space       = &unicode.RangeTable{} // Space
	Unary       = &unicode.RangeTable{} // Unary - operators that are only unary
	Vary        = &unicode.RangeTable{} // Vary - operators that can be unary or binary depending on context
	Special     = &unicode.RangeTable{} // Special - characters not covered by other classes
)

func getClass(s string) *unicode.RangeTable {
	switch s {
	case "N":
		return Normal
	case "A":
		return Alphabetic
	case "B":
		return Binary
	case "C":
		return Closing
	case "D":
		return Diacritic
	case "F":
		return Fence
	case "G":
		return Glyph_Part
	case "L":
		return Large
	case "O":
		return Opening
	case "P":
		return Punctuation
	case "R":
		return Relation
	case "S":
		return Space
	case "U":
		return Unary
	case "V":
		return Vary
	case "X":
		return Special
	}
	panic(fmt.Sprintf("support for class %q not yet implemented", s))
}

func addRange(class *unicode.RangeTable, raw_range string) error {
	if strings.Contains(raw_range, "..") {
		// range syntax:
		//
		//    0030..0039
		parts := strings.Split(raw_range, "..")
		if len(parts) != 2 {
			return errors.Errorf("expected two .. delimited parts in range, got %v", raw_range)
		}
		raw_start := parts[0]
		raw_end := parts[1] // end is inclusive
		start, err := strconv.ParseUint(raw_start, 16, 64)
		if err != nil {
			return errors.WithStack(err)
		}
		end, err := strconv.ParseUint(raw_end, 16, 64)
		if err != nil {
			return errors.WithStack(err)
		}
		switch {
		case start <= math.MaxUint16 && end <= math.MaxUint16:
			r16 := unicode.Range16{
				Lo: uint16(start),
				Hi: uint16(end),
				// TODO: set Stride?
			}
			class.R16 = append(class.R16, r16)
		default:
			r32 := unicode.Range32{
				Lo: uint32(start),
				Hi: uint32(end),
				// TODO: set Stride?
			}
			class.R32 = append(class.R32, r32)
		}
	} else {
		raw_num := raw_range
		// single code point syntax:
		//
		//    002F
		num, err := strconv.ParseUint(raw_num, 16, 32)
		if err != nil {
			return errors.WithStack(err)
		}
		switch {
		case num <= math.MaxUint16:
			r16 := unicode.Range16{
				Lo: uint16(num),
				Hi: uint16(num),
				// TODO: set Stride?
			}
			class.R16 = append(class.R16, r16)
		default:
			r32 := unicode.Range32{
				Lo: uint32(num),
				Hi: uint32(num),
				// TODO: set Stride?
			}
			class.R32 = append(class.R32, r32)
		}
	}
	return nil
}
