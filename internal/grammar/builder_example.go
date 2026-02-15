// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package grammar

import "fmt"

func ExampleBuilder() {
	// real code would implement the span function
	var span func() *Span

	b := NewBuilder("example.y")

	E := b.EnsureNonterminal("expr", span())
	PLUS := b.EnsureTerminal("PLUS", span())
	NUM := b.EnsureTerminal("NUM", span())

	b.DefinePrecedenceGroup(AssocLeft, []*Symbol{PLUS}, span())

	rb := b.BeginRule(E, span())
	rb.Alt([]*SymbolRef{
		b.NewRef(E, "A", span()),
		b.NewRef(PLUS, "", span()),
		b.NewRef(E, "B", span()),
	}, b.NewAction("...semantic code...", span()), nil, span())
	rb.Alt([]*SymbolRef{
		b.NewRef(NUM, "N", span()),
	}, nil, nil, span())
	rb.End()

	g := b.Finalize()
	_ = g
	if b.HasErrors() {
		for _, d := range b.Diagnostics() {
			fmt.Println(d.Error())
		}
		// real code would return the errors
		// return fmt.Errorf("grammar has errors")
	}
	// Safe to proceed to table construction, so real could would generate the tables
	// return generateTables(g)
}
