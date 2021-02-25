package bytecode

import (
	"math"
	"reflect"
	"strings"

	"github.com/tucats/ego/datatypes"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/util"
)

/******************************************\
*                                         *
*     M A T H   P R I M I T I V E S       *
*                                         *
\******************************************/

// NegateImpl instruction processor pops the top stack
// item and pushes it's negative. For booleans, this is
// a "not" operation; for numeric values it is simple
// negation. For an array, it reverses the order of the
// array elements.
func NegateImpl(c *Context, i interface{}) *errors.EgoError {
	v, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	// Cannot do math on a nil value
	if datatypes.IsNil(v) {
		return c.NewError(errors.InvalidTypeError)
	}

	switch value := v.(type) {
	case bool:
		_ = c.stackPush(!value)

	case int:
		_ = c.stackPush(-value)

	case float64:
		_ = c.stackPush(0.0 - value)

	case *datatypes.EgoArray:
		// Create an array in inverse order.
		r := datatypes.NewArray(value.ValueType(), value.Len())

		for n := 0; n < value.Len(); n = n + 1 {
			d, _ := value.Get(n)
			_ = r.Set(value.Len()-n-1, d)
		}

		_ = c.stackPush(r)

	case []interface{}:
		// Create an array in inverse order.
		r := make([]interface{}, len(value))

		for n, d := range value {
			r[len(value)-n-1] = d
		}

		_ = c.stackPush(r)

	default:
		return c.NewError(errors.InvalidTypeError)
	}

	return nil
}

// AddImpl bytecode instruction processor. This removes the top two
// items and adds them together. For boolean values, this is an OR
// operation. For numeric values, it is arithmetic addition. For
// strings or arrays, it concatenates the two items. For a struct,
// it merges the addend into the first struct.
func AddImpl(c *Context, i interface{}) *errors.EgoError {
	v2, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	v1, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	// Cannot do math on a nil value
	if datatypes.IsNil(v1) || datatypes.IsNil(v2) {
		return c.NewError(errors.InvalidTypeError)
	}

	switch vx := v1.(type) {
	case error:
		return c.stackPush(vx.Error() + util.GetString(v2))

		/*	case *errors.EgoError:
			_ = vx.Context(v2)

			return c.Push(vx) */

	// Is it a native array we are concatenating to?
	case []interface{}:
		switch vy := v2.(type) {
		// Array requires a deep concatenation.
		case []interface{}:
			// If we're in static type mode, each member of the
			// array being added must match the type of the target
			// array.
			if c.Static {
				arrayType := reflect.TypeOf(vx[0])

				for _, vv := range vy {
					if arrayType != reflect.TypeOf(vv) {
						return c.NewError(errors.InvalidTypeError)
					}
				}
			}

			newArray := append(vx, vy...)

			return c.stackPush(newArray)

		// Everything else is a simple append.
		default:
			newArray := append(vx, v2)

			return c.stackPush(newArray)
		}

		// You can add a map to another map.
	case map[string]interface{}:
		switch vy := v2.(type) {
		case map[string]interface{}:
			for k, v := range vy {
				vx[k] = v
			}

			return c.stackPush(vx)

		default:
			return c.NewError(errors.InvalidTypeError)
		}

		// All other types are scalar math.
	default:
		v1, v2 = util.Normalize(v1, v2)

		switch v1.(type) {
		case int:
			return c.stackPush(v1.(int) + v2.(int))

		case float64:
			return c.stackPush(v1.(float64) + v2.(float64))

		case string:
			return c.stackPush(v1.(string) + v2.(string))

		case bool:
			return c.stackPush(v1.(bool) && v2.(bool))

		default:
			return c.NewError(errors.InvalidTypeError)
		}
	}
}

// AndImpl bytecode instruction processor.
func AndImpl(c *Context, i interface{}) *errors.EgoError {
	v1, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	v2, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	// Cannot do math on a nil value
	if datatypes.IsNil(v1) || datatypes.IsNil(v2) {
		return c.NewError(errors.InvalidTypeError)
	}

	return c.stackPush(util.GetBool(v1) && util.GetBool(v2))
}

// OrImpl bytecode instruction processor.
func OrImpl(c *Context, i interface{}) *errors.EgoError {
	v1, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	v2, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	// Cannot do math on a nil value
	if datatypes.IsNil(v1) || datatypes.IsNil(v2) {
		return c.NewError(errors.InvalidTypeError)
	}

	return c.stackPush(util.GetBool(v1) || util.GetBool(v2))
}

// SubtractImpl instruction processor removes two items from the
// stack and subtracts them. For numeric values, this is arithmetic
// subtraction. For an array, the item to be subtracted is removed
// from the array (in any array location it is found).
func SubtractImpl(c *Context, i interface{}) *errors.EgoError {
	v2, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	v1, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	// Cannot do math on a nil value
	if datatypes.IsNil(v1) || datatypes.IsNil(v2) {
		return c.NewError(errors.InvalidTypeError)
	}

	switch vx := v1.(type) {
	// For an array, make a copy removing the item to be subtracted.
	case []interface{}:
		newArray := make([]interface{}, 0)

		for _, v := range vx {
			if !reflect.DeepEqual(v2, v) {
				newArray = append(newArray, v)
			}
		}

		return c.stackPush(newArray)

	// Everything else is a scalar subtraction.
	default:
		v1, v2 = util.Normalize(v1, v2)

		switch v1.(type) {
		case int:
			return c.stackPush(v1.(int) - v2.(int))

		case float64:
			return c.stackPush(v1.(float64) - v2.(float64))

		case string:
			s := strings.ReplaceAll(v1.(string), v2.(string), "")

			return c.stackPush(s)

		default:
			return c.NewError(errors.InvalidTypeError)
		}
	}
}

// MultiplyImpl bytecode instruction processor.
func MultiplyImpl(c *Context, i interface{}) *errors.EgoError {
	v2, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	v1, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	// Cannot do math on a nil value
	if datatypes.IsNil(v1) || datatypes.IsNil(v2) {
		return c.NewError(errors.InvalidTypeError)
	}

	v1, v2 = util.Normalize(v1, v2)

	switch v1.(type) {
	case int:
		return c.stackPush(v1.(int) * v2.(int))

	case float64:
		return c.stackPush(v1.(float64) * v2.(float64))

	case bool:
		return c.stackPush(v1.(bool) || v2.(bool))

	default:
		return c.NewError(errors.InvalidTypeError)
	}
}

// ExponentImpl bytecode instruction processor.
func ExponentImpl(c *Context, i interface{}) *errors.EgoError {
	v2, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	v1, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	v1, v2 = util.Normalize(v1, v2)

	// Cannot do math on a nil value
	if datatypes.IsNil(v1) || datatypes.IsNil(v2) {
		return c.NewError(errors.InvalidTypeError)
	}

	switch v1.(type) {
	case int:
		if v2.(int) == 0 {
			return c.stackPush(0)
		}

		if v2.(int) == 1 {
			return c.stackPush(v1)
		}

		prod := v1.(int)

		for n := 2; n <= v2.(int); n = n + 1 {
			prod = prod * v1.(int)
		}

		return c.stackPush(prod)

	case float64:
		return c.stackPush(math.Pow(v1.(float64), v2.(float64)))

	default:
		return c.NewError(errors.InvalidTypeError)
	}
}

// DivideImpl bytecode instruction processor.
func DivideImpl(c *Context, i interface{}) *errors.EgoError {
	if c.sp < 1 {
		return c.NewError(errors.StackUnderflowError)
	}

	v2, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	v1, err := c.Pop()
	if !errors.Nil(err) {
		return err
	}

	// Cannot do math on a nil value
	if datatypes.IsNil(v1) || datatypes.IsNil(v2) {
		return c.NewError(errors.InvalidTypeError)
	}

	v1, v2 = util.Normalize(v1, v2)

	switch v1.(type) {
	case int:
		if v2.(int) == 0 {
			return c.NewError(errors.DivisionByZeroError)
		}

		return c.stackPush(v1.(int) / v2.(int))

	case float64:
		if v2.(float64) == 0 {
			return c.NewError(errors.DivisionByZeroError)
		}

		return c.stackPush(v1.(float64) / v2.(float64))

	default:
		return c.NewError(errors.InvalidTypeError)
	}
}
