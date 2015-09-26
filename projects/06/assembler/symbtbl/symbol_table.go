package symbtbl

import "sync"

// SymbolTable is a map table of symbol strings and its addresses.
// SymbolTable is thread safe, so it can be used in multiple goroutines.
type SymbolTable struct {
	mu    sync.RWMutex
	m     map[string]uintptr
	vaddr uintptr // variable symbol's address
}

// NewSymbolTable returns a new SymbolTable initialized by tbl.
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		m:     map[string]uintptr{},
		vaddr: 0x10,
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
	st.mu.Lock()
	defer st.mu.Unlock()

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

// AddVar adds a variable symbol into st. The variable symbol's address is automatically set.
//
// A variable symbol's address is 16 (0x10) in the inital state,
// and every time a symbol is added it is automatically incremented by 1.
func (st *SymbolTable) AddVar(symb string) {
	st.AddEntry(symb, st.vaddr)

	st.mu.Lock()
	st.vaddr++
	st.mu.Unlock()
}
