package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	store := make(map[string]Symbol)
	return &SymbolTable{store: store}
}

func (st *SymbolTable) Define(identifier string) Symbol {
	s := Symbol{Name: identifier, Scope: GlobalScope, Index: st.numDefinitions}
	st.store[identifier] = s
	st.numDefinitions++

	return s
}

func (st *SymbolTable) Resolve(identifier string) (Symbol, bool) {
	sym, ok := st.store[identifier]
	return sym, ok
}
