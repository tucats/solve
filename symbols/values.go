package symbols

import "github.com/tucats/ego/app-cli/ui"

// A few notes about how values are stored in a symbol table.
//
// 1. There is a simple map in the symbol table that maps the
//    symbol name to a slot number. The slot number is a value
//    that increments each time a new symbol is created in that
//    table.
//
// 2. Slot numbers are used to locate the value itself. VAlues are
//    stored in a []interface{}, which is always created using the
//    SymbolAllocationSize value. The address of this array is then
//    stored in the symbol table's Values array, which is an array
//    of pointers to the value lists.
//
// 3. When a value array is exhausted (i.e. a slot number exceeds
//    its capacity), a new value array is created and it's adderess
//	  is added to the Values array list.
//
// This somewhat cumbersome mechanism guarantees that, for the life
// of a symbol table, any address-of-value operations remain valid
// even if the symbol table must be expanded due to value list
// exhaustion. The new value list is appended to the Values array,
// and if they weren't pointers to the value list, then the array
// memory manager subsystem might move the arrays, breaking the
// address-of values.

// For the current symbol table, allocate the initial values list and
// create the Values array with it's address. If the bin map is
// nil, create it as well.
func (s *SymbolTable) initializeValues() {
	if s.symbols == nil {
		s.symbols = map[string]*SymbolAttribute{}
	}

	if s.values == nil {
		bin := make([]interface{}, SymbolAllocationSize)
		s.values = make([]*[]interface{}, 1)
		s.values[0] = &bin
		s.size = 0
	}
}

// Given an index and a value, store the value in the Values list.
func (s *SymbolTable) SetValue(index int, v interface{}) {
	if index == NoSlot {
		return
	}

	// The index number is divided by the size of each value list to
	// determine which value list to use. The index number modulo the
	// maximum value list size gives the slot number within the selected
	// bin number.
	bin := index / SymbolAllocationSize
	for bin >= len(s.values) {
		newBin := make([]interface{}, SymbolAllocationSize)
		s.values = append(s.values, &newBin)

		ui.Log(ui.SymbolLogger, "%s, create new value bin", s.Name)
	}

	slot := index % SymbolAllocationSize
	(*s.values[bin])[slot] = v
}

// Given an index, retrieve a value from the Values list.
func (s *SymbolTable) GetValue(index int) interface{} {
	if index == NoSlot {
		return nil
	}

	bin := index / SymbolAllocationSize
	slot := index % SymbolAllocationSize

	if bin >= len(s.values) {
		return nil
	}

	return (*s.values[bin])[slot]
}

// Given an index, return the address of the value in that
// slot.
func (s *SymbolTable) AddressOfValue(index int) *interface{} {
	if index == NoSlot {
		return nil
	}

	bin := index / SymbolAllocationSize
	slot := index % SymbolAllocationSize

	if bin >= len(s.values) {
		return nil
	}

	return &(*s.values[bin])[slot]
}
