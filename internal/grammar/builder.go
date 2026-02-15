// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package grammar

import (
	"fmt"
	"strings"
)

// Diagnostic is a structured error/warning emitted during building/validation.
type Diagnostic struct {
	Level DiagnosticLevel
	Msg   string
	At    *Span
}

type DiagnosticLevel uint8

const (
	DiagError DiagnosticLevel = iota + 1
	DiagWarn
)

func (d Diagnostic) Error() string {
	if d.At == nil {
		return d.Msg
	}
	return fmt.Sprintf("%s:%d:%d: %s", d.At.File, d.At.Line, d.At.Column, d.Msg)
}

// Builder builds a Grammar incrementally, collecting diagnostics instead of
// failing hard. This is what your grammar-file parser should talk to.
type Builder struct {
	g *Grammar

	// precedenceCounter increments each time we see a precedence directive group.
	precedenceCounter int

	diags []Diagnostic
}

// NewBuilder creates a new Builder with an empty Grammar.
func NewBuilder(fileLabel string) *Builder {
	g := &Grammar{
		Name:          "",
		Start:         nil,
		Symbols:       nil,
		SymbolsByName: map[string]*Symbol{},
		Rules:         nil,
		Directives:    map[string]string{},
	}
	_ = fileLabel // kept for future defaults; spans carry filenames
	return &Builder{g: g}
}

// Grammar returns the built grammar (even if there are diagnostics).
func (b *Builder) Grammar() *Grammar { return b.g }

// Diagnostics returns all diagnostics collected so far.
func (b *Builder) Diagnostics() []Diagnostic { return append([]Diagnostic(nil), b.diags...) }

// HasErrors reports whether any error-level diagnostics exist.
func (b *Builder) HasErrors() bool {
	for _, d := range b.diags {
		if d.Level == DiagError {
			return true
		}
	}
	return false
}

// Lookup a *Symbol, true if `name` already exists in the symbol table.
// Otherwise, returns nil, false.
func (b *Builder) Lookup(name string) (*Symbol, bool) {
	// Intern trims whitespace, so we want to normalize here as well
	name = strings.TrimSpace(name)
	symbol, ok := b.g.SymbolsByName[name]
	return symbol, ok
}

func (b *Builder) error(at *Span, msg string, args ...any) {
	b.diags = append(b.diags, Diagnostic{
		Level: DiagError,
		Msg:   fmt.Sprintf(msg, args...),
		At:    at,
	})
}

func (b *Builder) warn(at *Span, msg string, args ...any) {
	b.diags = append(b.diags, Diagnostic{
		Level: DiagWarn,
		Msg:   fmt.Sprintf(msg, args...),
		At:    at,
	})
}

// ---------------------------
// Symbol interning & metadata
// ---------------------------

// Intern gets or creates a symbol with the given name and kind.
// If the symbol already exists with a different kind, we keep the original and
// record an error.
func (b *Builder) Intern(name string, kind SymbolKind, at *Span) *Symbol {
	name = strings.TrimSpace(name)
	if name == "" {
		b.error(at, "symbol name is empty")
		// Return a dummy symbol so callers can continue.
		return b.internDummy(at)
	}

	if sym, ok := b.g.SymbolsByName[name]; ok {
		if sym.Kind != kind {
			b.error(at, "symbol %q previously declared as %s, cannot redeclare as %s",
				name, kindString(sym.Kind), kindString(kind))
		}
		return sym
	}

	sym := &Symbol{
		ID:         SymbolID(len(b.g.Symbols)),
		Name:       name,
		Kind:       kind,
		TypeTag:    "",
		Precedence: 0,
		Assoc:      AssocNone,
		DeclaredAt: at,
	}
	b.g.Symbols = append(b.g.Symbols, sym)
	b.g.SymbolsByName[name] = sym
	return sym
}

// EnsureTerminal is a convenience for grammar parsers that see a token name.
func (b *Builder) EnsureTerminal(name string, at *Span) *Symbol {
	return b.Intern(name, SymTerminal, at)
}

// EnsureNonterminal is a convenience for grammar parsers that see an LHS name.
func (b *Builder) EnsureNonterminal(name string, at *Span) *Symbol {
	return b.Intern(name, SymNonterminal, at)
}

// SetTypeTag sets the type tag for a symbol (via %type or %token).
// If conflicting tags are applied, an error is recorded and the first wins.
func (b *Builder) SetTypeTag(sym *Symbol, typeTag string, at *Span) {
	if sym == nil {
		return
	}
	typeTag = strings.TrimSpace(typeTag)
	if typeTag == "" {
		return
	}
	if sym.TypeTag != "" && sym.TypeTag != typeTag {
		b.error(at, "symbol %q already has type %q; cannot set to %q", sym.Name, sym.TypeTag, typeTag)
		return
	}
	sym.TypeTag = typeTag
}

// SetStart sets the grammar start symbol.
func (b *Builder) SetStart(sym *Symbol, at *Span) {
	if sym == nil {
		return
	}
	if sym.Kind != SymNonterminal {
		b.error(at, "start symbol %q must be a nonterminal", sym.Name)
		return
	}
	if b.g.Start != nil && b.g.Start != sym {
		b.warn(at, "start symbol changed from %q to %q", b.g.Start.Name, sym.Name)
	}
	b.g.Start = sym
}

// ---------------------------
// Precedence directives
// ---------------------------

// DefinePrecedenceGroup applies a precedence/associativity to a list of terminals.
// Call this when your parser reads something like:
//
//	%left PLUS MINUS
//	%right POW
func (b *Builder) DefinePrecedenceGroup(assoc Assoc, terminals []*Symbol, at *Span) {
	if assoc == AssocNone {
		b.error(at, "precedence group must have associativity (left/right/nonassoc)")
		return
	}
	b.precedenceCounter++
	level := b.precedenceCounter

	for _, t := range terminals {
		if t == nil {
			continue
		}
		if t.Kind != SymTerminal {
			b.error(at, "precedence can only be assigned to terminals; %q is %s", t.Name, kindString(t.Kind))
			continue
		}
		// If precedence was set already, Lemon-ish tools usually allow it but it’s
		// often a mistake. We'll warn and keep the first.
		if t.Precedence != 0 {
			b.warn(at, "terminal %q already has precedence %d; ignoring new precedence %d", t.Name, t.Precedence, level)
			continue
		}
		t.Precedence = level
		t.Assoc = assoc
	}
}

// ---------------------------
// Rules & productions
// ---------------------------

// RuleBuilder helps construct one Rule with multiple alternatives.
// Typical usage:
//
//	rb := b.BeginRule(lhs, at)
//	rb.Alt([]*SymbolRef{...}, action, prec, at)
//	rb.Alt(...)
//	rb.End()
type RuleBuilder struct {
	b    *Builder
	rule *Rule
	done bool
}

// BeginRule starts a new rule for the given LHS.
func (b *Builder) BeginRule(lhs *Symbol, at *Span) *RuleBuilder {
	if lhs == nil {
		lhs = b.internDummy(at)
	}
	if lhs.Kind != SymNonterminal {
		b.error(at, "rule LHS %q must be a nonterminal", lhs.Name)
	}

	r := &Rule{LHS: lhs, Alternatives: nil, At: at}
	b.g.Rules = append(b.g.Rules, r)

	// If no explicit start symbol yet, infer from first rule (common behavior).
	if b.g.Start == nil && lhs.Kind == SymNonterminal {
		b.g.Start = lhs
	}

	return &RuleBuilder{b: b, rule: r}
}

// Alt adds an alternative to the current rule.
func (rb *RuleBuilder) Alt(rhs []*SymbolRef, action *Action, prec *Symbol, at *Span) {
	if rb == nil || rb.done || rb.rule == nil {
		return
	}

	// Validate RHS refs are not nil.
	for i, sr := range rhs {
		if sr == nil || sr.Sym == nil {
			rb.b.error(at, "rhs symbol at position %d is nil", i)
		}
	}

	if prec != nil && prec.Kind != SymTerminal {
		rb.b.error(at, "precedence override symbol %q must be a terminal", prec.Name)
		prec = nil
	}

	rb.rule.Alternatives = append(rb.rule.Alternatives, &Alternative{
		RHS:     rhs,
		Action:  action,
		PrecSym: prec,
		At:      at,
	})
}

// End marks the rule builder as finished (optional but helps prevent misuse).
func (rb *RuleBuilder) End() {
	if rb == nil {
		return
	}
	rb.done = true
}

// NewRef creates an RHS reference (with optional label).
func (b *Builder) NewRef(sym *Symbol, label string, at *Span) *SymbolRef {
	if sym == nil {
		sym = b.internDummy(at)
	}
	label = strings.TrimSpace(label)
	return &SymbolRef{Sym: sym, Label: label, At: at}
}

// NewAction creates an action block wrapper.
func (b *Builder) NewAction(raw string, at *Span) *Action {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return &Action{Raw: raw, At: at}
}

// ---------------------------
// Generic directives passthrough
// ---------------------------

// SetDirective stores a directive key/value pair for later stages.
func (b *Builder) SetDirective(key, value string, at *Span) {
	key = strings.TrimSpace(key)
	if key == "" {
		b.error(at, "directive key is empty")
		return
	}
	// Warn on overwrite; keep last.
	if _, exists := b.g.Directives[key]; exists {
		b.warn(at, "directive %q overwritten", key)
	}
	b.g.Directives[key] = value
}

// ---------------------------
// Helpers
// ---------------------------

func kindString(k SymbolKind) string {
	switch k {
	case SymTerminal:
		return "terminal"
	case SymNonterminal:
		return "nonterminal"
	default:
		return "unknown"
	}
}

func (b *Builder) internDummy(at *Span) *Symbol {
	// Ensure one stable dummy per builder (so errors don’t explode symbols).
	const name = "<invalid>"
	if sym, ok := b.g.SymbolsByName[name]; ok {
		return sym
	}
	return b.Intern(name, SymNonterminal, at)
}
