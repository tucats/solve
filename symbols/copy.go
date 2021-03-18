package symbols

import "github.com/google/uuid"

// Make a copy of the symbol table, retaining the same values
// as before (in fact, the values are shared between the tables).
// This is mostly used to create unique symbol and constant maps,
// which are needed for shallow clones of a compiler.
func (s *SymbolTable) Clone(withLock bool) *SymbolTable {
	if withLock {
		s.mutex.Lock()
		defer s.mutex.Unlock()
	}

	t := SymbolTable{
		Name:          s.Name,
		Package:       s.Package,
		Parent:        s.Parent,
		Symbols:       map[string]int{},
		Constants:     map[string]interface{}{},
		Values:        s.Values,
		ID:            uuid.New(),
		ValueSize:     s.ValueSize,
		ScopeBoundary: true,
		isRoot:        true,
	}

	for k, v := range s.Symbols {
		t.Symbols[k] = v
	}

	for k, v := range s.Constants {
		t.Constants[k] = v
	}

	return &t
}