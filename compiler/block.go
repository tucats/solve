package compiler

import (
	"github.com/tucats/ego/bytecode"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/tokenizer"
)

// compileBlock compiles a statement block. The leading { has already
// been parsed.
func (c *Compiler) compileBlock() error {
	parsing := true
	c.blockDepth++

	c.b.Emit(bytecode.PushScope)

	for parsing {
		if c.t.IsNext(tokenizer.BlockEndToken) {
			break
		}

		err := c.compileStatement()
		if err != nil {
			return err
		}

		// Skip over a semicolon if found
		_ = c.t.IsNext(tokenizer.SemicolonToken)

		if c.t.AtEnd() {
			return c.newError(errors.ErrMissingEndOfBlock)
		}
	}

	c.b.Emit(bytecode.PopScope)

	c.blockDepth--

	return nil
}

// Require that the next item be a block, enclosed in {} characters.
func (c *Compiler) compileRequiredBlock() error {
	// If an empty block, no work to do
	if c.t.IsNext(tokenizer.EmptyBlockToken) {
		return nil
	}

	// Otherwise, needs to start with the open block
	if !c.t.IsNext(tokenizer.BlockBeginToken) {
		return c.newError(errors.ErrMissingBlock)
	}

	// now compile and close the block.
	return c.compileBlock()
}
