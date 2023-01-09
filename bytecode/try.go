package bytecode

import (
	"github.com/tucats/ego/data"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/i18n"
)

const (
	// Discards the catch set, which means all errors are caught.
	AllErrorsCatchSet = 0

	// Set of errors that an ?optional is permitted to ignore.
	OptionalCatchSet = 1
)

var catchSets = [][]error{
	// OptionalCatchSet
	{
		errors.EgoError(errors.ErrUnknownMember),
		errors.EgoError(errors.ErrInvalidType),
		errors.EgoError(errors.ErrNilPointerReference),
		errors.EgoError(errors.ErrDivisionByZero),
		errors.EgoError(errors.ErrArrayIndex),
	},
}

// tryByteCode instruction processor.
func tryByteCode(c *Context, i interface{}) error {
	try := TryInfo{
		addr:    data.Int(i),
		catches: make([]error, 0),
	}
	c.tryStack = append(c.tryStack, try)

	return nil
}

// WillCatch instruction. This lets the code specify which errors
// are permitted to be caught; if the list is empty then all errors
// are caught.
func willCatchByteCode(c *Context, i interface{}) error {
	if len(c.tryStack) == 0 {
		return c.newError(errors.ErrTryCatchMismatch)
	}

	try := c.tryStack[len(c.tryStack)-1]
	if try.catches == nil {
		try.catches = make([]error, 0)
	}

	switch i := i.(type) {
	case int:
		if i > len(catchSets) {
			return c.newError(errors.ErrInternalCompiler).Context(i18n.E("invalid.catch.set",
				map[string]interface{}{"index": i}))
		}

		// Zero has a special meaning of "catch everything"
		if i == AllErrorsCatchSet {
			try.catches = make([]error, 0)
		} else {
			try.catches = append(try.catches, catchSets[i-1]...)
		}

	case *errors.EgoErrorMsg:
		try.catches = append(try.catches, i)

	case error:
		try.catches = append(try.catches, errors.EgoError(i))

	case string:
		try.catches = append(try.catches, errors.NewMessage(i))

	default:
		return c.newError(errors.ErrInvalidType).Context(data.TypeOf(i).String())
	}

	c.tryStack[len(c.tryStack)-1] = try

	return nil
}

// tryPopByteCode instruction processor.
func tryPopByteCode(c *Context, i interface{}) error {
	if len(c.tryStack) == 0 {
		return c.newError(errors.ErrTryCatchMismatch)
	}

	if len(c.tryStack) == 1 {
		c.tryStack = make([]TryInfo, 0)
	} else {
		c.tryStack = c.tryStack[:len(c.tryStack)-1]
	}

	_ = c.symbols.Delete("_error", true)

	return nil
}
