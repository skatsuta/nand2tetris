package symbtbl

import "sync"

// SymbolTable is a map table of symbol strings and its addresses.
// SymbolTable is thread safe, so it can be used in multiple goroutines.
type SymbolTable struct {
	m  map[string]uintptr
	mu sync.RWMutex
}

// NewSymbolTable returns a new SymbolTable initialized by tbl.
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		m: map[string]uintptr{},
	}
}

// AddEntry adds a pair (symb, addr) into the symbol table st.
func (st *SymbolTable) AddEntry(symb string, addr uintptr) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.m[symb] = addr
}

// AddEntries adds all the elements of m into the symbol table st.
func (st *SymbolTable) AddEntries(ent map[string]uintptr) {
	// copy the given map to ensure an internal map is not changed by external reference.
	for k, v := range ent {
		st.m[k] = v
	}
}

// Contains reports whether symb is contained in st.
func (st *SymbolTable) Contains(symb string) bool {
	st.mu.RLock()
	defer st.mu.RUnlock()

	_, found := st.m[symb]
	return found
}

// GetAddress returns an address corresponding to symb.
// If symb is not contained in st, it returns 0.
func (st *SymbolTable) GetAddress(symb string) uintptr {
	st.mu.RLock()
	defer st.mu.RUnlock()

	return st.m[symb]
}
