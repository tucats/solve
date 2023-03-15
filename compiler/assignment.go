package compiler

import (
	"github.com/tucats/ego/bytecode"
	"github.com/tucats/ego/data"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/tokenizer"
)

// compileAssignment compiles an assignment statement.
func (c *Compiler) compileAssignment() error {
	start := c.t.Mark()

	storeLValue, err := c.assignmentTarget()
	if err != nil {
		return err
	}

	// Check for auto-increment or decrement
	autoMode := bytecode.NoOperation

	if c.t.Peek(1) == tokenizer.IncrementToken {
		autoMode = bytecode.Add
	}

	if c.t.Peek(1) == tokenizer.DecrementToken {
		autoMode = bytecode.Sub
	}

	// If there was an auto increment/decrement, make sure the LValue is
	// a simple value. We can check this easily by ensuring the LValue store
	// code is only two instructions (which will always be a "store" followed
	// by a "drop to marker").
	if autoMode != bytecode.NoOperation {
		if storeLValue.Mark() > 2 {
			return c.error(errors.ErrInvalidAuto)
		}

		t := data.String(storeLValue.Instruction(0).Operand)

		c.b.Emit(bytecode.Load, t)
		c.b.Emit(bytecode.Push, 1)
		c.b.Emit(autoMode)
		c.b.Emit(bytecode.Dup)
		c.b.Emit(bytecode.Store, t)
		c.t.Advance(1)

		return nil
	}

	// Not auto-anything, so verify that this is a legit assignment
	if !c.t.AnyNext(tokenizer.DefineToken,
		tokenizer.AssignToken,
		tokenizer.ChannelReceiveToken,
		tokenizer.AddAssignToken,
		tokenizer.SubtractAssignToken,
		tokenizer.MultiplyAssignToken,
		tokenizer.DivideAssignToken) {
		return c.error(errors.ErrMissingAssignment)
	}

	if c.t.AnyNext(tokenizer.SemicolonToken, tokenizer.EndOfTokens) {
		return c.error(errors.ErrMissingExpression)
	}

	// Handle implicit operators, line += or /= which do both
	// a math operation and an assignment.
	mode := bytecode.NoOperation

	switch c.t.Peek(0) {
	case tokenizer.AddAssignToken:
		mode = bytecode.Add

	case tokenizer.SubtractAssignToken:
		mode = bytecode.Sub

	case tokenizer.MultiplyAssignToken:
		mode = bytecode.Mul

	case tokenizer.DivideAssignToken:
		mode = bytecode.Div
	}

	// If we found an explicit operation, then let's do it.
	if mode != bytecode.NoOperation {
		// Back the tokenizer up to the lvalue token, because we need to
		// re-parse this as an expression instead of an lvalue to get one
		// of the terms for the operation.
		c.t.Set(start)

		e1, err := c.Expression()
		if err != nil {
			return err
		}

		// Parse over the operator again.
		if !c.t.AnyNext(tokenizer.AddAssignToken,
			tokenizer.SubtractAssignToken,
			tokenizer.MultiplyAssignToken,
			tokenizer.DivideAssignToken) {
			return errors.ErrMissingAssignment
		}

		// And then parse the second term that follows the implicit operator.
		e2, err := c.Expression()
		if err != nil {
			return err
		}

		// Emit the expressions and the operator, and then store using the code
		// generated for the lvalue.
		c.b.Append(e1)
		c.b.Append(e2)
		c.b.Emit(mode)
		c.b.Append(storeLValue)

		return nil
	}

	// If this is a construct like   x := <-ch   skip over the :=
	_ = c.t.IsNext(tokenizer.ChannelReceiveToken)

	// Seems like a simple assignment at this point, so parse the expression
	// to be assigned, emit the code for that expression, and then emit the code
	// that will store the result in the lvalue.
	expressionCode, err := c.Expression()
	if err != nil {
		return err
	}

	c.b.Append(expressionCode)

	// If this assignment was an interfaace{} unwrap operation, then
	// we need to modify the lvalue store by removing the last bytecode
	// (which is a DropToMarker). Then add code that checks to see if there
	// was abandoned info on the stack that should trigger an error if false.
	if c.flags.hasUnwrap {
		if storeLValue.StoreCount() < 2 {
			storeLValue.Remove(storeLValue.Mark() - 1)
			c.b.Emit(bytecode.Swap)
			c.b.Append(storeLValue)
			c.b.Emit(bytecode.IfError, errors.ErrTypeMismatch)
			c.b.Emit(bytecode.DropToMarker, bytecode.NewStackMarker("let"))

			return nil
		} else {
			c.b.Emit(bytecode.Swap)
		}
	}

	c.b.Append(storeLValue)

	return nil
}
