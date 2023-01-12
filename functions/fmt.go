package functions

import (
	"fmt"
	"strings"

	"github.com/tucats/ego/data"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/symbols"
	"github.com/tucats/ego/tokenizer"
	"github.com/tucats/ego/util"
)

// Printf implements fmt.printf() and is a wrapper around the native Go function.
func Printf(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	length := 0

	str, err := Sprintf(s, args)
	if err == nil {
		length, _ = fmt.Printf("%s", data.String(str))
	}

	return length, err
}

// Sprintf implements fmt.sprintf() and is a wrapper around the native Go function.
func Sprintf(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return 0, nil
	}

	fmtString := data.String(args[0])

	if len(args) == 1 {
		return fmtString, nil
	}

	return fmt.Sprintf(fmtString, args[1:]...), nil
}

// Print implements fmt.Print() and is a wrapper around the native Go function.
func Print(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	var b strings.Builder

	for i, v := range args {
		if i > 0 {
			b.WriteString(" ")
		}

		b.WriteString(FormatAsString(s, v))
	}

	text, e2 := fmt.Printf("%s", b.String())

	if e2 != nil {
		e2 = errors.NewError(e2)
	}

	return text, e2
}

// Println implements fmt.Println() and is a wrapper around the native Go function.
func Println(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	var b strings.Builder

	for i, v := range args {
		if i > 0 {
			b.WriteString(" ")
		}

		b.WriteString(FormatAsString(s, v))
	}

	text, e2 := fmt.Printf("%s\n", b.String())

	if e2 != nil {
		e2 = errors.NewError(e2)
	}

	return text, e2
}

// FormatAsString will attempt to use the String() function of the
// object type passed in, if it is a typed struct.  Otherwise, it
// just returns the Unquoted format value.
func FormatAsString(s *symbols.SymbolTable, v interface{}) string {
	if m, ok := v.(*data.Struct); ok {
		if f := m.GetType().Function("String"); f != nil {
			if fmt, ok := f.(func(s *symbols.SymbolTable, args []interface{}) (interface{}, error)); ok {
				local := symbols.NewChildSymbolTable("local to format", s)
				local.SetAlways("__this", v)

				if si, err := fmt(local, []interface{}{}); err == nil {
					if str, ok := si.(string); ok {
						return str
					}
				}
			}
		}
	}

	return data.FormatUnquoted(v)
}

func Sscanf(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	dataString := data.String(args[0])
	formatString := data.String(args[1])

	// Verify the remaining arguments are all pointers, and unwrap them.
	pointerList := make([]*interface{}, len(args)-2)

	for i, v := range args[2:] {
		if data.TypeOfPointer(v).IsUndefined() {
			return nil, errors.ErrNotAPointer
		}

		if content, ok := v.(*interface{}); ok {
			pointerList[i] = content
		}
	}

	// Do the scan, returning an array of values
	items, err := scanner(dataString, formatString)
	if err != nil {
		return 0, errors.NewError(err).Context("Sscanf()")
	}

	// Stride over the return value pointers, assigning as many
	// items as we got.
	for idx, p := range pointerList {
		if idx >= len(items) {
			break
		}

		*p = items[idx]
	}

	return len(items), nil
}

func scanner(data, format string) ([]interface{}, error) {
	var err error

	result := make([]interface{}, 0)

	fTokens := tokenizer.New(format)
	dTokens := tokenizer.New(data)
	d := dTokens.Tokens
	f := []string{}
	parsingVerb := false

	// Scan over the token, collapsing format verbs into a
	// single token.
	for _, token := range fTokens.Tokens {
		if parsingVerb {
			// Add to the previous token
			f[len(f)-1] = f[len(f)-1] + token.Spelling()

			// Are we at the end of a supported format string?
			if util.InList(token.Spelling(), "b", "x", "o", "s", "t", "f", "d", "v") {
				parsingVerb = false
			}
		} else {
			f = append(f, token.Spelling())
			if token == tokenizer.ModuloToken {
				parsingVerb = true
			}
		}
	}

	parsing := true

	// Now scan over the format tokens, which now represent either
	// required tokens in the input data or format operations.
	for idx, token := range f {
		if !parsing {
			break
		}

		if token[:1] == "%" {
			switch token[len(token)-1:] {
			case "v":
				var v interface{}

				_, e := fmt.Sscanf(d[idx].Spelling(), token, &v)
				if e != nil {
					err = errors.NewError(e).Context("Sscanf()")
					parsing = false

					break
				}

				result = append(result, v)

			case "s":
				v := ""

				_, e := fmt.Sscanf(d[idx].Spelling(), token, &v)
				if e != nil {
					err = errors.NewError(e).Context("Sscanf()")
					parsing = false

					break
				}

				result = append(result, v)

			case "t":
				v := false

				_, e := fmt.Sscanf(d[idx].Spelling(), token, &v)
				if e != nil {
					err = errors.NewError(e).Context("Sscanf()")
					parsing = false

					break
				}

				result = append(result, v)

			case "f":
				v := 0.0

				_, e := fmt.Sscanf(d[idx].Spelling(), token, &v)
				if e != nil {
					err = errors.NewError(e).Context("Sscanf()")
					parsing = false

					break
				}

				result = append(result, v)

			case "d", "b", "x", "o":
				v := 0

				_, e := fmt.Sscanf(d[idx].Spelling(), token, &v)
				if e != nil {
					err = errors.NewError(e).Context("Sscanf()")
					parsing = false

					break
				}

				result = append(result, v)
			}
		} else {
			if token != d[idx].Spelling() {
				parsing = false
			}
		}
	}

	return result, err
}
