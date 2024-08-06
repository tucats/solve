package compiler

import (
	"github.com/tucats/ego/bytecode"
	"github.com/tucats/ego/data"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/tokenizer"
)

// Compile an initializer, given a type definition. This can be a literal
// initializer in braces or a simple value.
func (c *Compiler) compileInitializer(t *data.Type) error {
	if !c.t.IsNext(tokenizer.DataBeginToken) {
		// It's not an initializer constant, but it could still be an expression. Try the
		// top-level expression compiler.
		return c.conditional()
	}

	base := t
	if t.IsTypeDefinition() {
		base = t.BaseType()
	}

	switch base.Kind() {
	case data.StructKind:
		count := 0

		// It's possible this is an initializer of an ordered list of values. If so, they must
		// match the number of fields, and are stored in the designated order. We determine this
		// is the case by checking for an expression followed by a comma or the end of the data

		tokenMark := c.t.Mark()

		if _, err := c.Expression(); err == nil && c.t.Peek(1) == tokenizer.CommaToken || c.t.Peek(1) == tokenizer.DataEndToken {
			// First, back up the tokenizer position
			c.t.Set(tokenMark)

			fieldNames := base.FieldNames()

			for !c.t.IsNext(tokenizer.DataEndToken) {
				if count >= len(fieldNames) {
					return c.error(errors.ErrInitializerCount, count)
				}

				// Parse the next value expression
				fieldName := fieldNames[count]
				fieldType, _ := base.Field(fieldName)

				// Generate the initializer value from the expression
				if err := c.compileInitializer(fieldType); err != nil {
					return err
				}

				// Now emit the name of the field
				c.b.Emit(bytecode.Push, fieldName)

				count++

				if c.t.IsNext(tokenizer.DataEndToken) {
					break
				}

				if !c.t.IsNext(tokenizer.CommaToken) {
					return c.error(errors.ErrInvalidList)
				}
			}

			if count < len(fieldNames) {
				return c.error(errors.ErrInitializerCount, count)
			}

		} else {
			// First, back up the tokenizer position
			c.t.Set(tokenMark)

			// Scane the list of field names and values
			for !c.t.IsNext(tokenizer.DataEndToken) {
				// Pairs of name:value
				name := c.t.Next()
				if !name.IsIdentifier() {
					return c.error(errors.ErrInvalidSymbolName)
				}

				name = tokenizer.NewIdentifierToken(c.normalize(name.Spelling()))

				fieldType, err := base.Field(name.Spelling())
				if err != nil {
					return err
				}

				if !c.t.IsNext(tokenizer.ColonToken) {
					return c.error(errors.ErrMissingColon)
				}

				if err = c.compileInitializer(fieldType); err != nil {
					return err
				}

				// Now emit the name (names always come first on the stack)
				c.b.Emit(bytecode.Push, name)

				count++

				if c.t.IsNext(tokenizer.DataEndToken) {
					break
				}

				if !c.t.IsNext(tokenizer.CommaToken) {
					return c.error(errors.ErrInvalidList)
				}
			}
		}

		c.b.Emit(bytecode.Push, t)
		c.b.Emit(bytecode.Push, data.TypeMDKey)
		c.b.Emit(bytecode.Struct, count+1)

		return nil

	case data.MapKind:
		count := 0

		for !c.t.IsNext(tokenizer.DataEndToken) {
			// Pairs of values with a colon between.
			if err := c.unary(); err != nil {
				return err
			}

			c.b.Emit(bytecode.Coerce, base.KeyType())

			if !c.t.IsNext(tokenizer.ColonToken) {
				return c.error(errors.ErrMissingColon)
			}

			// Note we compile the value using ourselves, to allow for nested
			// type specifications.
			if err := c.compileInitializer(base.BaseType()); err != nil {
				return err
			}

			count++

			if c.t.IsNext(tokenizer.DataEndToken) {
				break
			}

			if !c.t.IsNext(tokenizer.CommaToken) {
				return c.error(errors.ErrInvalidList)
			}
		}

		c.b.Emit(bytecode.Push, base.BaseType())
		c.b.Emit(bytecode.Push, base.KeyType())
		c.b.Emit(bytecode.MakeMap, count)

		return nil

	case data.ArrayKind:
		count := 0

		for !c.t.IsNext(tokenizer.DataEndToken) {
			// Values separated by commas.
			if err := c.compileInitializer(base.BaseType()); err != nil {
				return err
			}

			count++

			if c.t.IsNext(tokenizer.DataEndToken) {
				break
			}

			if !c.t.IsNext(tokenizer.CommaToken) {
				return c.error(errors.ErrInvalidList)
			}
		}

		// Emit the type as the final datum, and then emit the instruction
		// that will construct the array from the type and stack items.
		c.b.Emit(bytecode.Push, base.BaseType())
		c.b.Emit(bytecode.MakeArray, count)

		return nil

	default:
		if err := c.unary(); err != nil {
			return err
		}

		// If we are doing dynamic typing, let's allow a coercion as well.
		c.b.Emit(bytecode.Coerce, base)

		return nil
	}
}
