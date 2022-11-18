package runtime

import (
	"strings"
	"sync"

	"github.com/tucats/ego/app-cli/tables"
	"github.com/tucats/ego/app-cli/ui"
	"github.com/tucats/ego/compiler"
	"github.com/tucats/ego/datatypes"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/symbols"
)

var tableTypeDef *datatypes.Type
var tableTypeDefLock sync.Mutex

func initTableTypeDef() {
	tableTypeDefLock.Lock()
	defer tableTypeDefLock.Unlock()

	if tableTypeDef == nil {
		t, _ := compiler.CompileTypeSpec(tableTypeSpec)

		t.DefineFunctions(map[string]interface{}{
			"AddRow":     TableAddRow,
			"Close":      TableClose,
			"Sort":       TableSort,
			"Print":      TablePrint,
			"Format":     TableFormat,
			"Align":      TableAlign,
			"String":     TableString,
			"Pagination": TablePagination,
		})

		tableTypeDef = t
	}
}

// TableNew implements the New() table package function. This accepts a list
// of column names (as individual arguments or an array of strings) and allocates
// a new table. Additionally, the column names can contain alignment information;
// a name with a leading ":" is left-aligned, and a trailing ":" is right-
// aligned. In either case the ":" is removed from the name.
func TableNew(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	if len(args) == 0 {
		return nil, errors.New(errors.ErrArgumentCount)
	}

	// Fetch the arguments as column headings. If the value is passed by array,
	// extract each array member as a column name.
	headings := []string{}

	for _, h := range args {
		if list, ok := h.(*datatypes.EgoArray); ok {
			for idx := 0; idx < list.Len(); idx++ {
				str, _ := list.Get(idx)
				headings = append(headings, datatypes.GetString(str))
			}
		} else if list, ok := h.([]interface{}); ok {
			for _, hh := range list {
				headings = append(headings, datatypes.GetString(hh))
			}
		} else {
			headings = append(headings, datatypes.GetString(h))
		}
	}

	// Scan over the heading strings and look for alignment cues. If found,
	// remove the ":" cue character, and record the specified (or default)
	// alignment for each column.
	align := make([]int, len(headings))

	for i := 0; i < len(headings); i = i + 1 {
		h := headings[i]
		if strings.HasPrefix(h, ":") && strings.HasSuffix(h, ":") {
			align[i] = tables.AlignmentCenter
			headings[i] = strings.TrimPrefix(strings.TrimSuffix(h, ":"), ":")
		} else if strings.HasPrefix(h, ":") {
			align[i] = tables.AlignmentLeft
			headings[i] = strings.TrimPrefix(h, ":")
		} else if strings.HasSuffix(h, ":") {
			align[i] = tables.AlignmentRight
			headings[i] = strings.TrimSuffix(h, ":")
		} else {
			align[i] = tables.AlignmentLeft
		}
	}

	// Create the new table object, and set the alignment for each column heading now.
	t, err := tables.New(headings)
	if !errors.Nil(err) {
		return nil, err
	}

	for i, v := range align {
		_ = t.SetAlignment(i, v)
	}

	// Move the string array of headings into a native array type, which can
	// be read by the caller.
	headingsArray := datatypes.NewArray(&datatypes.StringType, len(headings))

	for i, h := range headings {
		_ = headingsArray.Set(i, h)
	}

	initTableTypeDef()

	result := datatypes.NewStruct(tableTypeDef)
	result.SetAlways(tableFieldName, t)
	result.SetAlways(headingsFieldName, headingsArray)
	result.SetReadonly(true)

	return result, nil
}

// TableClose closes the table handle, and releases any memory resources
// being held by the table.
func TableClose(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	_, err := getTable(s)
	if !errors.Nil(err) {
		return nil, err
	}

	this := getThisStruct(s)
	this.SetAlways(tableFieldName, nil)

	return true, err
}

// TableClose closes the table handle, and releases any memory resources
// being held by the table.
func TablePagination(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	if len(args) != 2 {
		return nil, errors.New(errors.ErrInvalidVariableArguments)
	}

	h := datatypes.GetInt(args[0])
	w := datatypes.GetInt(args[1])

	t, err := getTable(s)
	if !errors.Nil(err) {
		return nil, err
	}

	t.SetPagination(h, w)

	return true, err
}

// TableAddRow adds a row to the table. This can either be a list of values, or
// a struct. When it's a struct, each column name must match a struct member
// name, and the associated value is used as the table cell value. If a list of
// values is given, they are stored in the row in the same order that the columns
// were defined when the table was created.
func TableAddRow(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	t, err := getTable(s)
	if errors.Nil(err) {
		if len(args) > 0 {
			if m, ok := args[0].(*datatypes.EgoStruct); ok {
				if len(args) > 1 {
					err = errors.New(errors.ErrArgumentCount)
				} else {
					values := make([]string, len(m.FieldNames()))

					for _, k := range m.FieldNames() {
						v := m.GetAlways(k)
						if v == nil {
							return nil, errors.New(errors.ErrInvalidField)
						}

						p, ok := t.FindColumn(k)
						if ok {
							values[p] = datatypes.GetString(v)
						}
					}

					err = t.AddRow(values)
				}
			} else {
				if m, ok := args[0].([]interface{}); ok {
					if len(args) > 1 {
						err = errors.New(errors.ErrArgumentCount)

						return err, err
					}
					err = t.AddRowItems(m...)
				} else {
					err = t.AddRowItems(args...)
				}
			}
		}
	}

	return err, err
}

// TableSort sorts the rows of the table. If you specify multiple arguments
// (column names) the sort is performed in the reverse order specified; that
// is the least-significant sort is performed first, then the next-most-
// significant sort, etc. until the first argument, which is the most
// significant sort. The column names can start with a tilde ("~") character
// to reverse the sort order from it's default value of ascending to descending.
func TableSort(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	t, err := getTable(s)
	if errors.Nil(err) {
		for i := len(args) - 1; i >= 0; i = i - 1 {
			v := args[i]
			ascending := true

			heading := datatypes.GetString(v)
			if strings.HasPrefix(heading, "~") {
				ascending = false
				heading = heading[1:]
			}

			pos, found := t.FindColumn(heading)
			if !found {
				err = errors.New(errors.ErrInvalidColumnName).Context(heading)
			} else {
				err = t.SortRows(pos, ascending)
			}
		}
	}

	return err, err
}

// TableFormat specifies the headings format. It accepts two values, which
// are both booleans. The first indicates if a headings row is to be printed
// in the output. The second is examined only if the headings value is true;
// it controls whether an underline string is printed under the column names.
func TableFormat(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	if len(args) > 2 {
		err := errors.New(errors.ErrArgumentCount)

		return err, err
	}

	t, err := getTable(s)
	if errors.Nil(err) {
		headings := true
		lines := true

		if len(args) > 0 {
			headings = datatypes.GetBool(args[0])
			lines = headings
		}

		if len(args) > 1 {
			lines = datatypes.GetBool(args[1])
		}

		t.ShowHeadings(headings)
		t.ShowUnderlines(lines)
	}

	return err, err
}

// TableAlign specifies alignment for a given column.
func TableAlign(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	if len(args) > 2 {
		err := errors.New(errors.ErrArgumentCount)

		return err, err
	}

	t, err := getTable(s)
	if errors.Nil(err) {
		column := 0

		if columnName, ok := args[0].(string); ok {
			column, ok = t.FindColumn(columnName)
			if !ok {
				err = errors.New(errors.ErrInvalidColumnName).Context(columnName)

				return err, err
			}
		} else {
			column = datatypes.GetInt(args[0])
		}

		mode := tables.AlignmentLeft

		if modeName, ok := args[1].(string); ok {
			switch strings.ToLower(modeName) {
			case "left":
				mode = tables.AlignmentLeft

			case "right":
				mode = tables.AlignmentRight

			case "center":
				mode = tables.AlignmentCenter

			default:
				err = errors.New(errors.ErrAlignment).Context(modeName)

				return err, err
			}
		}

		err = t.SetAlignment(column, mode)
	}

	return err, err
}

// TablePrint prints a table to the default output, in the default --output-format
// type (text or json).
func TablePrint(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	fmt := ui.OutputFormat

	if len(args) > 0 {
		fmt = datatypes.GetString(args[0])
	}

	t, err := getTable(s)
	if errors.Nil(err) {
		err = t.Print(fmt)
	}

	return err, err
}

// TableString formats a table as a string in the default output.
func TableString(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	fmt := ui.OutputFormat

	if len(args) > 0 {
		fmt = datatypes.GetString(args[0])
	}

	t, err := getTable(s)
	if errors.Nil(err) {
		return t.String(fmt)
	}

	return nil, err
}

// getTable searches the symbol table for the client receiver ("__this")
// variable, validates that it contains a table object, and returns the
// native table object.
func getTable(symbols *symbols.SymbolTable) (*tables.Table, *errors.EgoError) {
	if g, ok := symbols.Get("__this"); ok {
		if gc, ok := g.(*datatypes.EgoStruct); ok {
			if tbl, ok := gc.Get(tableFieldName); ok {
				if tp, ok := tbl.(*tables.Table); ok {
					if tp == nil {
						return nil, errors.New(errors.ErrTableClosed)
					}

					return tp, nil
				}
			}
		}
	}

	return nil, errors.New(errors.ErrNoFunctionReceiver)
}
