// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/maloquacious/semver"
)

var (
	version = semver.Version{
		Minor:      1,
		PreRelease: "alpha",
	}
)

func main() {
	// Parse command-line flags similar to the original lemon tool.
	// For reference only.
	var (
		// Basic options
		baseFlagPtr       = flag.Bool("b", false, "Show only the basis in report")
		noCompressFlagPtr = flag.Bool("c", false, "Don't compress the action table")
		outputDirPtr      = flag.String("d", "", "Output directory")
		showHelpPtr       = flag.Bool("?", false, "Show help")
		showVersionPtr    = flag.Bool("x", false, "Show version")
		statsFlagPtr      = flag.Bool("s", false, "Show statistics about table generation")
		templateFilePtr   = flag.String("T", "", "Specify a template file")

		// Advanced options
		definePtr          = flag.String("D", "", "Define an %ifdef macro")
		makeheadersPtr     = flag.Bool("m", false, "Output a makeheaders compatible file")
		noLineNosPtr       = flag.Bool("l", false, "Do not print #line statements")
		printGrammarPtr    = flag.Bool("g", false, "Print grammar without actions")
		printPreprocessPtr = flag.Bool("E", false, "Print input file after preprocessing")
		quietPtr           = flag.Bool("q", false, "Don't print the report file")
		noResortPtr        = flag.Bool("r", false, "Do not sort or renumber states")
		showPrecedencePtr  = flag.Bool("p", false, "Show precedence levels in the report")
		sqlPtr             = flag.Bool("S", false, "Generate an SQLite3 table of parser statistics")

		// Debug options
		debugPtr = flag.Bool("debug", false, "Enable debug output during parser generation")
		tracePtr = flag.Bool("trace", false, "Enable trace output in the generated parser")
	)

	flag.Parse()

	if *showHelpPtr {
		fmt.Println("Guanabana LALR(1) Parser Generator")
		flag.PrintDefaults()
		return
	}

	if *showVersionPtr {
		fmt.Printf("Guanabana Parser Generator Version %s\n", version.String())
		return
	}

	// Check if a grammar file was specified
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Error: No grammar file specified")
		fmt.Println("Usage: guanabana [options] grammar-file")
		os.Exit(1)
	}

	grammarFile := args[0]

	// Create a new parser and process the grammar file
	p := Parser{}

	// Basic options
	p.Basisflag = *baseFlagPtr
	p.NoResort = *noCompressFlagPtr
	p.Stats = *statsFlagPtr
	p.TemplateFile = *templateFilePtr

	// Use the 'generated' directory by default to avoid cluttering with C files
	if *outputDirPtr == "" {
		p.Outdir = "generated"
	} else {
		p.Outdir = *outputDirPtr
	}

	// Advanced options
	if *definePtr != "" {
		// In the original Lemon, this defines a preprocessing macro
		// We'll store them and pass to our grammar preprocessor when implemented
		// For now we'll just print a warning
		fmt.Printf("Warning: -D option not fully implemented yet\n")
	}
	p.MakeHeaders = *makeheadersPtr
	p.NoLineNos = *noLineNosPtr
	p.PrintGrammar = *printGrammarPtr
	p.PrintPreprocess = *printPreprocessPtr
	p.Quiet = *quietPtr
	p.NoResort = *noResortPtr
	p.ShowPrecedence = *showPrecedencePtr
	p.SQL = *sqlPtr

	// Debug options
	p.Debug = *debugPtr
	p.Trace = *tracePtr
	err := p.GenerateParser(grammarFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// Parser is a copy of the original lemon parser generator. It's not used; it
// is here for reference purposes. We'll replace it with the Guanabana Parser.
// It contains the entire state of the parser generator, including the grammar,
// configuration options, and the parser state machine. This is the main type
// used to parse grammar files and generate parser code.
type Parser struct {
	// Parser configuration
	Basisflag      bool   // Output only basis configurations
	NoResort       bool   // Do not sort or renumber states
	ShowPrecedence bool   // Show precedence conflicts in the report
	Quiet          bool   // Don't print non-essential information
	Stats          bool   // Print performance statistics
	Grammar        string // Input grammar file name
	StartRule      string // Name of the start rule
	IncludePath    string // Directory for inclusion preprocessor
	Outdir         string // Directory where files are written
	TemplateFile   string // Template file

	// Advanced options
	MakeHeaders     bool // Output a makeheaders compatible file
	NoLineNos       bool // Do not print #line statements
	PrintGrammar    bool // Print grammar without actions
	PrintPreprocess bool // Print input file after preprocessing
	SQL             bool // Generate an SQLite3 table of parser statistics

	// Debug options
	Debug bool // Enable debug output during parser generation
	Trace bool // Enable trace output in the generated parser

	//// Parser state
	//Nrule          int      // Number of rules
	//Nsymbol        int      // Number of symbols
	//Nstate         int      // Number of states
	//Rule           []*Rule  // Array of rules
	//Symbols        []*Symbol // Array of symbols
	//StartSym       *Symbol // Start symbol
	//TokenPrec      int     // Precedence for token symbols
	//ErrorSym       *Symbol // The error symbol
	//WildcardSym    *Symbol // The wildcard symbol
	//Name           string  // Name of the generated parser
	//TokenPrefix    string  // Prefix for token names
	//TokenType      string  // Type of terminal symbols
	//Vartype        string  // The default value of VARTYPE
	//StateSet       *StateSet // The set of states in the LALR(1) state machine

	// Directive values
	IncludeCode       string // Code from %include directive
	ExtraCode         string // Code from %code directive
	ExtraArgument     string // Type from %extra_argument directive
	TokenDestructor   string // Code from %token_destructor directive
	DefaultDestructor string // Code from %default_destructor directive
	SyntaxError       string // Code from %syntax_error directive
	ParseAccept       string // Code from %parse_accept directive
	ParseFailure      string // Code from %parse_failure directive
	StackOverflow     string // Code from %stack_overflow directive
	StackSize         int    // Value from %stack_size directive (default: 100)

	// Path names
	TemplateFilename string // The template file name
	TemplateContent  string // Content of the template file
	OutputFilename   string // The output file name
	HeaderFilename   string // The header file name
	ReportFilename   string // The report file name
}

func (p Parser) GenerateParser(grammarFile string) error {
	panic("not implemented")
}
