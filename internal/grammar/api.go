// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package grammar

// Sink is what the grammar-file parser talks to.
// It’s basically an event receiver.
type Sink interface {
	// ----- Directives -----
	// Examples: %start_symbol, %token, %type, %include, %default_type, etc.
	Directive(d Directive)

	// ----- Rules -----
	// Called when a new rule begins: LHS ::= ...
	BeginRule(lhs SymRef)

	// Called once per alternative:
	//   LHS ::= rhs1 rhs2 .   (or with action / precedence override)
	Alternative(alt Alt)

	// Called when the rule ends (often at '.' in Lemon-like syntax).
	EndRule(at *Span)

	// ----- Diagnostics shortcut -----
	// If the parser finds a syntax error, it can report it here.
	ParserError(at *Span, msg string)
}

// Directive is a structured directive record.
// Keep it generic: different directives have different payloads.
type Directive struct {
	Kind DirectiveKind
	At   *Span

	// Most directives are (key, value) or (key, list).
	Key   string
	Value string
	List  []SymRef
}

type DirectiveKind uint8

const (
	DirUnknown DirectiveKind = iota
	DirStartSymbol
	DirTokenType // e.g. %token_type {TYPE} or %token_type TYPE (depending on your format)
	DirType      // e.g. %type NonTerminal {TYPE}
	DirToken     // e.g. %token TOKENNAME
	DirLeft
	DirRight
	DirNonassoc
	DirInclude
	DirCode
	DirFallback
	// Add more as you meet them; parser stays the same shape.
)

// SymRef is how the parser refers to a symbol occurrence without needing
// builder internals. The builder decides whether it’s terminal/nonterminal
// based on context, spelling conventions, or explicit directives.
type SymRef struct {
	Name string
	At   *Span

	// Optional per-occurrence label (expr(A)) and/or type tag.
	Label   string
	TypeTag string
}

// Alt is a full alternative for the current rule.
type Alt struct {
	At *Span

	// RHS symbols in order
	RHS []SymRef

	// Optional action code block
	Action *Action

	// Optional precedence override (e.g. [PLUS] or %prec PLUS)
	Prec *SymRef
}
