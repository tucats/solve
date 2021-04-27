package compiler

import (
	bc "github.com/tucats/ego/bytecode"
	"github.com/tucats/ego/errors"
)

// conditional handles parsing the ?: trinary operator. The first term is
// converted to a boolean value, and if true the second term is returned, else
// the third term. All terms must be present.
func (c *Compiler) conditional() *errors.EgoError {
	// Parse the conditional
	err := c.relations()
	if !errors.Nil(err) {
		return err
	}

	// If this is not a conditional, we're done. Conditionals
	// are only permitted when extensions are enabled.
	if c.t.AtEnd() || !c.extensionsEnabled || c.t.Peek(1) != "?" {
		return nil
	}

	m1 := c.b.Mark()

	c.b.Emit(bc.BranchFalse, 0)

	// Parse both parts of the alternate values
	c.t.Advance(1)

	err = c.relations()
	if !errors.Nil(err) {
		return err
	}

	if c.t.AtEnd() || c.t.Peek(1) != ":" {
		return c.newError(errors.ErrMissingColon)
	}

	m2 := c.b.Mark()

	c.b.Emit(bc.Branch, 0)
	_ = c.b.SetAddressHere(m1)
	c.t.Advance(1)

	err = c.relations()
	if !errors.Nil(err) {
		return err
	}

	// Patch up the forward references.
	_ = c.b.SetAddressHere(m2)

	return nil
}
