package compiler

import (
	"github.com/tucats/ego/bytecode"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/tokenizer"
	"github.com/tucats/ego/util"
)

// isAssignmentTarget peeks ahead to see if this is likely to be an lValue
// object. This is used in cases where the parser might be in an
// otherwise ambiguous state.
func (c *Compiler) isAssignmentTarget() bool {
	// Remember were we are, and set it back when done.
	mark := c.t.Mark()
	defer c.t.Set(mark)

	// If this is a leading asterisk, that's fine.
	if c.t.Peek(1) == "*" {
		c.t.Advance(1)
	}

	// See if it's a symbol
	if name := c.t.Peek(1); !tokenizer.IsSymbol(name) {
		return false
	} else {
		// See if it's a reserved word.
		if tokenizer.IsReserved(name, c.extensionsEnabled) {
			return false
		}
	}

	// Let's look ahead to see if it contains any of the tell-tale
	// characters that indicate an lvalue. This does not say if it
	// is a valid/correct lvalue. We also stop searching at some point.
	for i := 2; i < 100; i = i + 1 {
		t := c.t.Peek(i)
		if util.InList(t, tokenizer.AssignToken, "=", "<-", "+=", "-=", "*=", "/=") {
			return true
		}

		// Is this an auto increment?
		if c.t.Peek(i) == "++" {
			return true
		}

		// Is this an auto decrement?
		if c.t.Peek(i) == "--" {
			return true
		}

		if tokenizer.IsReserved(t, c.extensionsEnabled) {
			return false
		}

		if util.InList(t, tokenizer.BlockBeginToken, tokenizer.SemicolonToken, tokenizer.EndOfTokens) {
			return false
		}
	}

	return false
}

// Check to see if this is a list of lvalues, which can occur
// in a multi-part assignment.
func assignmentTargetList(c *Compiler) (*bytecode.ByteCode, *errors.EgoError) {
	bc := bytecode.New("lvalue list")
	count := 0

	savedPosition := c.t.TokenP
	isLvalueList := false

	bc.Emit(bytecode.StackCheck, 1)

	if c.t.Peek(1) == "*" {
		return nil, c.newError(errors.ErrInvalidSymbolName, "*")
	}

	for {
		name := c.t.Next()
		if !tokenizer.IsSymbol(name) {
			c.t.Set(savedPosition)

			return nil, c.newError(errors.ErrInvalidSymbolName, name)
		}

		name = c.normalize(name)
		needLoad := true

		// Until we get to the end of the lvalue...
		for util.InList(c.t.Peek(1), ".", "[") {
			if needLoad {
				bc.Emit(bytecode.Load, name)

				needLoad = false
			}

			err := c.lvalueTerm(bc)
			if !errors.Nil(err) {
				return nil, err
			}
		}

		// Cheating here a bit; this opcode does an optional create
		// if it's not found anywhere in the tree already.
		bc.Emit(bytecode.SymbolOptCreate, name)
		patchStore(bc, name, false, false)

		count++

		if c.t.Peek(1) == tokenizer.CommaToken {
			c.t.Advance(1)

			isLvalueList = true

			continue
		}

		if util.InList(c.t.Peek(1), "=", tokenizer.AssignToken, "<-") {
			break
		}
	}

	if isLvalueList {
		// TODO if this is a channel store, then a list is not supported yet.
		if c.t.Peek(1) == "<-" {
			return nil, c.newError(errors.ErrInvalidChannelList)
		}

		// Patch up the stack size check. We can use the SetAddress
		// operator to do this because it really just updates the
		// integer instruction argument.
		_ = bc.SetAddress(0, count)

		// Also, add an instruction that will drop the marker value
		bc.Emit(bytecode.DropToMarker)

		return bc, nil
	}

	c.t.TokenP = savedPosition

	return nil, c.newError(errors.ErrNotAnLValueList)
}

// assignmentTarget compiles the information on the left side of
// an assignment. This information is used later to store the
// data in the named object.
func (c *Compiler) assignmentTarget() (*bytecode.ByteCode, *errors.EgoError) {
	if bc, err := assignmentTargetList(c); errors.Nil(err) {
		return bc, nil
	}

	// Add a marker in the regular code stream here
	c.b.Emit(bytecode.Push, bytecode.NewStackMarker("let"))

	bc := bytecode.New("lvalue")
	isPointer := false

	name := c.t.Next()
	if name == "*" {
		isPointer = true
		name = c.t.Next()
	}

	if !tokenizer.IsSymbol(name) {
		return nil, c.newError(errors.ErrInvalidSymbolName, name)
	}

	name = c.normalize(name)
	needLoad := true

	// Until we get to the end of the lvalue...
	for c.t.Peek(1) == "." || c.t.Peek(1) == "[" {
		if needLoad {
			bc.Emit(bytecode.Load, name)

			needLoad = false
		}

		err := c.lvalueTerm(bc)
		if !errors.Nil(err) {
			return nil, err
		}
	}

	// Quick optimization; if the name is "_" it just means
	// discard and we can shortcircuit that.
	if name == bytecode.DiscardedVariableName {
		bc.Emit(bytecode.Drop, 1)
	} else {
		// If its the case of x := <-c  then skip the assignment
		if util.InList(c.t.Peek(1), "=", tokenizer.AssignToken) && c.t.Peek(2) == "<-" {
			c.t.Advance(1)
		}
		if c.t.Peek(1) == tokenizer.AssignToken {
			bc.Emit(bytecode.SymbolCreate, name)
		}

		patchStore(bc, name, isPointer, c.t.Peek(1) == "<-")
	}

	bc.Emit(bytecode.DropToMarker, bytecode.NewStackMarker("let"))

	return bc, nil
}

// Helper function for LValue processing. If the token stream we are
// generating ends in a LoadIndex, but this is the last part of the
// storagebytecode, convert the last operation to a Store which writes
// the value back.
func patchStore(bc *bytecode.ByteCode, name string, isPointer, isChan bool) {
	// Is the last operation in the stack referecing
	// a parent object? If so, convert the last one to
	// a store operation.
	ops := bc.Opcodes()

	opsPos := bc.Mark() - 1
	if opsPos > 0 && ops[opsPos].Operation == bytecode.LoadIndex && ops[opsPos].Operand == nil {
		ops[opsPos] = bytecode.Instruction{Operation: bytecode.StoreIndex, Operand: nil}
	} else {
		if isChan {
			bc.Emit(bytecode.StoreChan, name)
		} else {
			if isPointer {
				bc.Emit(bytecode.StoreViaPointer, name)
			} else {
				bc.Emit(bytecode.Store, name)
			}
		}
	}
}

// lvalueTerm parses secondary lvalue operations (array indexes, or struct member dereferences).
func (c *Compiler) lvalueTerm(bc *bytecode.ByteCode) *errors.EgoError {
	term := c.t.Peek(1)
	if term == "[" {
		c.t.Advance(1)

		ix, err := c.Expression()
		if !errors.Nil(err) {
			return err
		}

		bc.Append(ix)

		if !c.t.IsNext("]") {
			return c.newError(errors.ErrMissingBracket)
		}

		bc.Emit(bytecode.LoadIndex)

		return nil
	}

	if term == "." {
		c.t.Advance(1)

		member := c.t.Next()
		if !tokenizer.IsSymbol(member) {
			return c.newError(errors.ErrInvalidSymbolName, member)
		}

		// Must do this as a push/loadindex in case the struct is
		// actuall a typed struct.
		bc.Emit(bytecode.Push, c.normalize(member))
		bc.Emit(bytecode.LoadIndex)

		return nil
	}

	return nil
}
