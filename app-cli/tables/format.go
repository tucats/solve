package tables

import (
	"strings"

	"github.com/tucats/ego/errors"
)

// RowLimit sets the row limit for output (<0 means all rows).
func (t *Table) RowLimit(limit int) *Table {
	if limit <= 0 {
		t.rowLimit = -1
	} else {
		t.rowLimit = limit
	}

	return t
}

// ShowUnderlines enables underlining of column headings when the parameter is true.
func (t *Table) ShowUnderlines(flag bool) *Table {
	t.showUnderlines = flag

	return t
}

// ShowHeadings disables printing of column headings when the parameter is true.
func (t *Table) ShowHeadings(flag bool) *Table {
	t.showHeadings = flag

	return t
}

// ShowRowNumbers enables printing of column headings when the parameter is true.
func (t *Table) ShowRowNumbers(flag bool) *Table {
	t.showRowNumbers = flag

	return t
}

// SetMinimumWidth specifies the minimum width of a column. The column number is
// always zero-based.
func (t *Table) SetMinimumWidth(n int, w int) *errors.EgoError {
	if n < 0 || n >= t.columnCount {
		return errors.New(errors.InvalidColumnNumberError).Context(n)
	}

	if w < 0 {
		return errors.New(errors.InvalidColumnWidthError).Context(w)
	}

	if w > t.maxWidth[n] {
		t.maxWidth[n] = w
	}

	return nil
}

// SetStartingRow specifies the first row of the table to be
// printed. A value less than zero is an error.
func (t *Table) SetStartingRow(s int) *errors.EgoError {
	if s < 1 {
		return errors.New(errors.InvalidRowNumberError).Context(s)
	}

	t.startingRow = s - 1

	return nil
}

// SetSpacing specifies the spaces between columns in output.
func (t *Table) SetSpacing(s int) *errors.EgoError {
	if s < 0 {
		return errors.New(errors.InvalidSpacingError).Context(s)
	}

	var buffer strings.Builder

	for i := 0; i < s; i++ {
		buffer.WriteRune(' ')
	}

	t.spacing = buffer.String()

	return nil
}

// SetIndent specifies the spaces to indent each heading and row.
func (t *Table) SetIndent(s int) *errors.EgoError {
	var buffer strings.Builder

	if s < 0 {
		return errors.New(errors.InvalidSpacingError).Context(s)
	}

	for i := 0; i < s; i++ {
		buffer.WriteRune(' ')
	}

	t.indent = buffer.String()

	return nil
}

// SetAlignment sets the alignment for a given column.
func (t *Table) SetAlignment(column int, alignment int) *errors.EgoError {
	if column < 0 || column >= t.columnCount {
		return errors.New(errors.InvalidColumnNumberError).Context(column)
	}

	switch alignment {
	case AlignmentLeft:
		t.alignment[column] = AlignmentLeft

	case AlignmentRight:
		t.alignment[column] = AlignmentRight

	case AlignmentCenter:
		t.alignment[column] = AlignmentCenter

	default:
		return errors.New(errors.InvalidAlignmentError).Context(alignment)
	}

	return nil
}
