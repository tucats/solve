package tokenizer

import (
	"strconv"
	"strings"
	"text/scanner"
	"unicode"

	"github.com/tucats/ego/app-cli/ui"
)

// Tokenizer is an instance of a tokenized string.
type Tokenizer struct {
	Source []string
	Tokens []Token
	TokenP int
	Line   []int
	Pos    []int
}

// EndOfTokens is a reserved token that means end of the buffer was reached.
var EndOfTokens = Token{class: EndOfTokensClass}

// ToTheEnd means to advance the token stream to the end.
const ToTheEnd = 999999

// This describes a token that is "crushed"; that is converting a sequence
// of tokens as generated by the standard Go scanner into a single Ego token
// where appropriate.
type crushedToken struct {
	source []Token
	result Token
}

// This is the table of tokens that are "crushed" into a single token.
var crushedTokens = []crushedToken{
	{
		source: []Token{AddToken, AssignToken},
		result: AddAssignToken,
	},
	{
		source: []Token{SubtractToken, AssignToken},
		result: SubtractAssignToken,
	},
	{
		source: []Token{MultiplyToken, AssignToken},
		result: MultiplyAssignToken,
	},
	{
		source: []Token{DivideToken, AssignToken},
		result: DivideAssignToken,
	},
	{
		source: []Token{AddToken, AddToken},
		result: IncrementToken,
	},
	{
		source: []Token{SubtractToken, SubtractToken},
		result: DecrementToken,
	},
	{
		source: []Token{InterfaceToken, DataBeginToken, DataEndToken},
		result: EmptyInterfaceToken,
	},
	{
		source: []Token{NewIdentifierToken(InterfaceToken.spelling), DataBeginToken, DataEndToken},
		result: EmptyInterfaceToken,
	},
	{
		source: []Token{BlockBeginToken, BlockEndToken},
		result: EmptyBlockToken,
	},
	{
		source: []Token{DotToken, DotToken, DotToken},
		result: VariadicToken,
	},
	{
		source: []Token{LessThanToken, SubtractToken},
		result: ChannelReceiveToken,
	},
	{
		source: []Token{GreaterThanToken, AssignToken},
		result: GreaterThanOrEqualsToken,
	},
	{
		source: []Token{LessThanToken, AssignToken},
		result: LessThanOrEqualsToken,
	},
	{
		source: []Token{AssignToken, AssignToken},
		result: EqualsToken,
	},
	{
		source: []Token{NotToken, AssignToken},
		result: NotEqualsToken,
	},
	{
		source: []Token{ColonToken, AssignToken},
		result: DefineToken,
	},
	{
		source: []Token{AndToken, AndToken},
		result: BooleanAndToken,
	},
	{
		source: []Token{OrToken, OrToken},
		result: BooleanOrToken,
	},
	{
		source: []Token{LessThanToken, LessThanToken},
		result: ShiftLeftToken,
	},
	{
		source: []Token{GreaterThanToken, GreaterThanToken},
		result: ShiftRightToken,
	},
}

// New creates a tokenizer instance and breaks the string
// up into an array of tokens.
func New(src string) *Tokenizer {
	var s scanner.Scanner

	t := Tokenizer{Source: splitLines(src), TokenP: 0}
	t.Tokens = make([]Token, 0)

	s.Init(strings.NewReader(src))
	s.Error = func(s *scanner.Scanner, msg string) { /* suppress messaging */ }
	s.Filename = "Input"

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		nextTokenSpelling := s.TokenText()

		var nextToken Token

		if TypeTokens[NewTypeToken(nextTokenSpelling)] {
			nextToken = NewTypeToken(nextTokenSpelling)
		} else if tx := NewReservedToken(nextTokenSpelling); tx.IsReserved(true) {
			nextToken = tx
		} else if IsSymbol(nextTokenSpelling) {
			nextToken = NewIdentifierToken(nextTokenSpelling)
		} else if SpecialTokens[NewSpecialToken(nextTokenSpelling)] {
			nextToken = NewSpecialToken(nextTokenSpelling)
		} else if strings.HasPrefix(nextTokenSpelling, "\"") && strings.HasSuffix(nextTokenSpelling, "\"") {
			rawString := unQuote(nextTokenSpelling)
			nextToken = NewStringToken(rawString)
		} else if strings.HasPrefix(nextTokenSpelling, "`") && strings.HasSuffix(nextTokenSpelling, "`") {
			nextToken = NewStringToken(strings.TrimPrefix(strings.TrimSuffix(nextTokenSpelling, "`"), "`"))
		} else {
			nextToken = Token{class: ValueTokenClass, spelling: nextTokenSpelling}
		}

		t.Tokens = append(t.Tokens, nextToken)
		column := s.Column

		// See if this is one of the special cases convert multiple tokens into
		// a single token?
		for _, crush := range crushedTokens {
			if len(crush.source) > len(t.Tokens) {
				continue
			}

			found := true
			// See if the current token stream now ends with a sequence that should
			// be collapsed. If we look at each source token and never get a mismatch
			// we know this was still found.
			for i, ch := range crush.source {
				if t.Tokens[len(t.Tokens)-len(crush.source)+i] != ch {
					found = false

					break
				}
			}

			// If we found a match here, lop off the individual tokens
			// and replace the "current" token with the crushed value
			if found {
				t.Tokens = append(t.Tokens[:len(t.Tokens)-len(crush.source)], crush.result)

				// We also must adjust the Line and Pos arrays accordingly. Remove as many
				// items from the end as needed.
				t.Line = t.Line[:len(t.Line)-len(crush.source)+1]
				t.Pos = t.Pos[:len(t.Pos)-len(crush.source)+1]

				// Adjust the column to reflect the character position of the
				// start of the crushed token.
				column = column - len(crush.result.Spelling())

				break
			}
		}

		// Add in the line from the scan and the (possibly adjusted) column
		t.Line = append(t.Line, s.Line)
		t.Pos = append(t.Pos, column)
	}

	if ui.IsActive(ui.TokenLogger) {
		ui.WriteLog(ui.TokenLogger, "Tokenizer contents:")

		for index, token := range t.Tokens {
			ui.WriteLog(ui.TokenLogger, "  [%2d:%2d] %v", t.Line[index], t.Pos[index], token)
		}
	}

	return &t
}

// Remainder returns the rest of the source, as initially presented to the
// tokenizer, from the current token position. This allows the caller to get
// "the rest" of a command line or other element as needed. If the token
// position is invalid (i.e. past end-of-tokens, for example) then an empty
// string is returned.
func (t *Tokenizer) Remainder() string {
	if t.TokenP < 0 || t.TokenP >= len(t.Pos) {
		return ""
	}

	p := t.Pos[t.TokenP] - 1
	s := t.GetSource()

	if p < 0 || p >= len(s) {
		return ""
	}

	return strings.TrimSuffix(s[p:], "\n")
}

// Next gets the next token in the tokenizer.
func (t *Tokenizer) Next() Token {
	if t.TokenP >= len(t.Tokens) {
		return EndOfTokens
	}

	token := t.Tokens[t.TokenP]
	t.TokenP++

	return token
}

// Next gets the next token in the tokenizer and returns it's
// text value as a string.
func (t *Tokenizer) NextText() string {
	if t.TokenP >= len(t.Tokens) {
		return EndOfTokens.spelling
	}

	token := t.Tokens[t.TokenP]
	t.TokenP++

	return token.spelling
}

// Peek looks ahead at the next token without advancing the pointer.
func (t *Tokenizer) Peek(offset int) Token {
	position := t.TokenP + (offset - 1)
	if position >= len(t.Tokens) || position < 0 {
		return EndOfTokens
	}

	return t.Tokens[position]
}

// Peek looks ahead at the next token without advancing the pointer.
func (t *Tokenizer) PeekText(offset int) string {
	pos := t.TokenP + (offset - 1)
	if pos >= len(t.Tokens) {
		return EndOfTokens.spelling
	}

	return t.Tokens[pos].spelling
}

// AtEnd indicates if we are at the end of the string.
func (t *Tokenizer) AtEnd() bool {
	return t.TokenP >= len(t.Tokens)
}

// Advance moves the pointer.
func (t *Tokenizer) Advance(p int) {
	t.TokenP = t.TokenP + p
	if t.TokenP < 0 {
		t.TokenP = 0
	} else if t.TokenP > len(t.Tokens) {
		t.TokenP = len(t.Tokens)
	}
}

// IsNext tests to see if the next token is the given token, and if so
// advances and returns true, else does not advance and returns false.
func (t *Tokenizer) IsNext(test Token) bool {
	if t.Peek(1) == test {
		t.Advance(1)

		return true
	}

	return false
}

// AnyNext tests to see if the next token is in the given  list
// of tokens, and if so  advances and returns true, else does not
// advance and returns false.
func (t *Tokenizer) AnyNext(test ...Token) bool {
	n := t.Peek(1)
	for _, v := range test {
		if n == v {
			t.Advance(1)

			return true
		}
	}

	return false
}

// IsSymbol is a utility function to determine if a string contains is a symbol name.
func IsSymbol(s string) bool {
	for n, c := range s {
		if c == '_' || unicode.IsLetter(c) {
			continue
		}

		if n > 0 && unicode.IsDigit(c) {
			continue
		}

		return false
	}

	return true
}

// GetLine returns a given line of text from the token stream.
// This actuals refers to the original line splits done when the
// source was first received.
func (t *Tokenizer) GetLine(line int) string {
	if line < 1 || line > len(t.Source) {
		return ""
	}

	return t.Source[line-1]
}

// splitLines splits a string by line endings, and returns the
// source as an array of strings.
func splitLines(src string) []string {
	// Are we seeing Windows-style line endings? If so, use that as
	// the split boundary.
	if strings.Index(src, "\r\n") > 0 {
		return strings.Split(src, "\r\n")
	}

	// Otherwise, simple split by new-line works fine.
	return strings.Split(src, "\n")
}

// GetSource returns the entire string of the tokenizer.
func (t *Tokenizer) GetSource() string {
	result := strings.Builder{}

	for _, line := range t.Source {
		result.WriteString(line)
		result.WriteRune('\n')
	}

	return result.String()
}

// GetTokens returns a string representing the tokens
// within the given range of tokens.
func (t *Tokenizer) GetTokens(pos1, pos2 int, spacing bool) string {
	p1 := pos1
	if p1 < 0 {
		p1 = 0
	} else if p1 > len(t.Tokens) {
		p1 = len(t.Tokens)
	}

	p2 := pos2
	if p2 < p1 {
		p2 = p1
	} else {
		if p2 > len(t.Tokens) {
			p2 = len(t.Tokens)
		}
	}

	var s strings.Builder

	for _, t := range t.Tokens[p1:p2] {
		s.WriteString(t.spelling)

		if spacing {
			s.WriteRune(' ')
		}
	}

	return s.String()
}

// InList is a support function that checks to see if a string matches
// any of a list of other strings.
func InList(s Token, test ...Token) bool {
	for _, t := range test {
		if s == t {
			return true
		}
	}

	return false
}

// unQuote will remove quotes and also process any escapes in the string.
// It is a wrapper around strconv.Unquote but does not report an error if
// the string is improperly formed.
func unQuote(input string) string {
	result, err := strconv.Unquote(input)
	if err != nil {
		return input
	}

	return result
}
