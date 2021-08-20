package compiler

import (
	"github.com/tucats/ego/bytecode"
	"github.com/tucats/ego/datatypes"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/tokenizer"
)

// compileIf compiles conditional statments. The verb is already
// removed from the token stream.
func (c *Compiler) compileIf() *errors.EgoError {
	if c.t.AnyNext(";", tokenizer.EndOfTokens) {
		return c.newError(errors.ErrMissingExpression)
	}

	// Compile the conditional expression
	bc, err := c.Expression()
	if !errors.Nil(err) {
		return err
	}

	c.b.Emit(bytecode.Push, datatypes.BoolType)
	c.b.Append(bc)
	c.b.Emit(bytecode.Call, 1)

	b1 := c.b.Mark()

	c.b.Emit(bytecode.BranchFalse, 0)

	// Compile the statement to be executed if true
	err = c.compileRequiredBlock()
	if !errors.Nil(err) {
		return err
	}

	// If there's an else clause, branch around it.
	if c.t.IsNext("else") {
		b2 := c.b.Mark()

		c.b.Emit(bytecode.Branch, 0)
		_ = c.b.SetAddressHere(b1)

		err = c.compileRequiredBlock()
		if !errors.Nil(err) {
			return err
		}

		_ = c.b.SetAddressHere(b2)
	} else {
		_ = c.b.SetAddressHere(b1)
	}

	return nil
}
