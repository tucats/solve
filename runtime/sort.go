package runtime

import (
	"sort"

	"github.com/tucats/ego/bytecode"
	"github.com/tucats/ego/datatypes"
	"github.com/tucats/ego/defs"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/symbols"
)

// sortSlice implements the sort.Slice() function. Beause this function requires a callback
// function written as bytecode, it cannot be in the functions package to avoid an import
// cycle problem. So this function (and others like it) are declared outside the functions
// package here in the runtime package, and are manually added to the dictionary when the
// run command is invoked.
func sortSlice(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, errors.EgoError(errors.ErrArgumentCount)
	}

	array, ok := args[0].(*datatypes.EgoArray)
	if !ok {
		return nil, errors.EgoError(errors.ErrArgumentType)
	}

	fn, ok := args[1].(*bytecode.ByteCode)
	if !ok {
		return nil, errors.EgoError(errors.ErrArgumentType)
	}

	var funcError error

	// Create a symbol table to use for the slice comparator callback function.
	sliceSymbols := symbols.NewChildSymbolTable("sort slice", s)

	// Coerce the name of the bytecode to represent that it is the
	// anonymous compare function value. We only do this if it is
	// actually anonymous.
	if fn.Name() == "" {
		fn.SetName(defs.Anon)
	}

	// Reusable context that will handle each callback.
	ctx := bytecode.NewContext(sliceSymbols, fn)

	// Use the native sort.Slice function, and provide a comparitor function
	// whose job is to run the supplied bytecode instructions, passing in
	// the two native arguments
	sort.Slice(array.BaseArray(), func(i, j int) bool {
		// Set the i,j variables as the current function arguments
		sliceSymbols.SetAlways("__args", datatypes.NewArrayFromArray(&datatypes.IntType, []interface{}{i, j}))

		// Run the comparator function
		if err := ctx.RunFromAddress(0); err != nil {
			if funcError == nil {
				funcError = err
			}

			return false
		}

		// Return the result as this function's value.
		return datatypes.GetBool(ctx.Result())
	})

	return array, funcError
}
