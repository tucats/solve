// Package expressions is a simple expression evaluator. It supports
// a rudementary symbol table with scoping, and knows about four data
// types (string, integer, double, and boolean). It does type casting as
// need automatically.
//
// The general pattern of use is:
//
//    e := expressions.New().WithText("expression string")
//    v, err := e.Eval(symbols.SymbolTable)
//
//	If the expression is to be evaluated only once, then you can simplify
//	the evaluation to:
//
//    v, err := expressions.Evaluate("expr string", *symbols.SymbolTable)
//
//  The value is returned as an opaque interface{} type. You can use the
//  following helper functions to retrieve the value from the interface,
//  and coerce the type if possible.
//
//    i := GetInt(v)
//    f := GetFloat(v)
//    s := GetString(v)
//    b := GetBool(v)
//
package expressions

import (
	"github.com/tucats/ego/bytecode"
	"github.com/tucats/ego/compiler"
	"github.com/tucats/ego/symbols"
	"github.com/tucats/ego/tokenizer"
)

// Expression is the type for an instance of the expresssion evaluator.
type Expression struct {
	t   *tokenizer.Tokenizer
	b   *bytecode.ByteCode
	c   bool
	err error
}

// New creates a new Expression object. The expression to evaluate is
// provided.
func New() *Expression {

	// Start with an Expression structure.
	e := &Expression{}

	return e
}

// WithNormalization expresses whether case normalization is to be used
// with this expression. This must be chained before any of the WithText,
// WithCompiler, or WithTokens calls.
func (e *Expression) WithNormalization(b bool) *Expression {
	e.c = b
	return e
}

// WithText provides the text for a compilation.
func (e *Expression) WithText(expr string) *Expression {
	// Create a compiler object and attach the tokenized expression
	cx := compiler.New().WithTokens(tokenizer.New(expr))
	cx.LowercaseIdentifiers = e.c

	// compile the code, store the generated bytecode and the
	// error, if any.
	e.b, e.err = cx.Expression()
	return e
}

// WithTokenizer creates a new Expression object. The expression to evaluate is
// provided.
func (e *Expression) WithTokenizer(t *tokenizer.Tokenizer) *Expression {

	cx := compiler.New()
	cx.LowercaseIdentifiers = e.c

	// tokenized already, just attach in progress
	e.t = t

	// compile
	e.b, e.err = cx.Expression()

	return e

}

// WithBytecode allocates an expression object and
// attaches the provided bytecode structure.
func (e *Expression) WithBytecode(b *bytecode.ByteCode) *Expression {
	e.b = b
	return e
}

// Error returns the last error seen on the expression object.
func (e *Expression) Error() error {
	return e.err
}

// Disasm calls the bytecode disassembler.
func (e *Expression) Disasm() {
	e.b.Disasm()
}

// GetBytecode returns the active bytecode for the expression
func (e *Expression) GetBytecode() *bytecode.ByteCode {
	return e.b
}

// Evaluate is a helper function for the case where a string is to
// be evaluated once and the value returned.
func Evaluate(expr string, s *symbols.SymbolTable) (interface{}, error) {
	return New().WithText(expr).Eval(s)
}
