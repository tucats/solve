package compiler

import (
	"github.com/tucats/ego/bytecode"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/tokenizer"
)

// compileFunctionCall handles the call statement. This is really the same as
// invoking a function in an expression, except there is no
// result value.
func (c *Compiler) compileFunctionCall() error {
	// Is this really panic, handled elsewhere?
	if c.flags.extensionsEnabled && c.t.Peek(0) == tokenizer.PanicToken {
		return c.compilePanic()
	}

	// Let's peek ahead to see if this is a legit function call. If the next token is
	// not an identifier, and it's not followed by a parenthesis or dot-notation identifier,
	// then this is not a function call and we're done.
	if !c.t.Peek(1).IsIdentifier() || (c.t.Peek(2) != tokenizer.StartOfListToken && c.t.Peek(2) != tokenizer.DotToken) {
		return c.error(errors.ErrInvalidFunctionCall)
	}

	// Parse the function as an expression. Place a marker on the stack before emitting
	// the expression for the call, so we can later discard unused return values.
	c.b.Emit(bytecode.Push, bytecode.NewStackMarker("call"))

	bc, err := c.Expression()
	if err != nil {
		return err
	}

	c.b.Append(bc)

	// We don't care about the result values, so flush to the marker.
	c.b.Emit(bytecode.DropToMarker, bytecode.NewStackMarker("call"))

	return nil
}
