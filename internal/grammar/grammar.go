// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package grammar

// SymbolID is a stable, dense ID (0..N-1) assigned during symbol interning.
type SymbolID int

// SymbolKind distinguishes terminals vs nonterminals.
type SymbolKind uint8

const (
	SymTerminal SymbolKind = iota + 1
	SymNonterminal
)

// Assoc is operator associativity for precedence handling.
type Assoc uint8

const (
	AssocNone Assoc = iota
	AssocLeft
	AssocRight
	AssocNonassoc
)

// Symbol is a named grammar symbol (terminal or nonterminal).
// Terminals and nonterminals share the same namespace in Lemon-style grammars.
type Symbol struct {
	ID   SymbolID
	Name string
	Kind SymbolKind

	// TypeTag is optional. Used by %type / %token_type-style directives.
	// We keep it as a string so you can carry "int", "Expr*", etc. without caring.
	TypeTag string

	// Precedence is 0 if none assigned; higher means binds tighter.
	Precedence int
	Assoc      Assoc

	// DeclaredAt is optional, but very useful for error messages.
	DeclaredAt *Span
}

// Grammar is the in-memory representation of a grammar file.
type Grammar struct {
	// Name is optional (some formats support it; handy for diagnostics/output).
	Name string

	// Start is the start symbol. If nil, it will be inferred later (typically
	// the LHS of the first rule), or set via a directive.
	Start *Symbol

	// Symbols is the intern table. SymbolsByName maps for quick lookup.
	Symbols       []*Symbol
	SymbolsByName map[string]*Symbol

	// Rules in source order.
	Rules []*Rule

	// Directives captures extra settings we don't want to hardcode yet.
	Directives map[string]string
}

// Rule is a production group: LHS ::= RHS1 | RHS2 | ...
type Rule struct {
	LHS          *Symbol
	Alternatives []*Alternative
	At           *Span
}

// Alternative is one RHS option for a rule.
type Alternative struct {
	// RHS is the sequence of symbols on the right-hand side.
	RHS []*SymbolRef

	// Action is an optional semantic action block (opaque to us for now).
	// We store raw text; later lessons can parse out argument names, etc.
	Action *Action

	// PrecSym optionally overrides the precedence/associativity of this
	// alternative. If nil, precedence is usually derived from the rightmost
	// terminal with precedence, depending on your rules.
	PrecSym *Symbol

	At *Span
}

// SymbolRef is a reference to a symbol in an RHS, with optional label/alias.
// Lemon-like grammars allow naming RHS terms for use in actions. Even if we
// don’t implement the action language, tracking labels helps diagnostics.
type SymbolRef struct {
	Sym *Symbol

	// Label is an optional name attached to this occurrence (e.g. "expr(A)").
	Label string

	At *Span
}

// Action is an opaque semantic action block associated with an alternative.
type Action struct {
	// Raw includes the text inside the braces (or however the grammar denotes it).
	Raw string
	At  *Span
}

// Span identifies a location in the source grammar file for diagnostics.
type Span struct {
	File string
	// 1-based, inclusive positions.
	Line   int
	Column int
	// Optional end position (can be zeroed if you only track a point).
	EndLine   int
	EndColumn int
}
