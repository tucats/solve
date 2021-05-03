package bytecode

import (
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"github.com/tucats/ego/app-cli/persistence"
	"github.com/tucats/ego/app-cli/ui"
	"github.com/tucats/ego/defs"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/symbols"
	"github.com/tucats/ego/tokenizer"
	"github.com/tucats/ego/util"
)

type This struct {
	name  string
	value interface{}
}

type TryInfo struct {
	addr    int
	catches []*errors.EgoError
}

// This value is updated atomically during context creation.
var nextThreadID int32 = 0

// Context holds the runtime information about an instance of bytecode being
// executed.
type Context struct {
	Name                 string
	bc                   *ByteCode
	symbols              *symbols.SymbolTable
	tokenizer            *tokenizer.Tokenizer
	stack                []interface{}
	tryStack             []TryInfo
	rangeStack           []*Range
	timerStack           []time.Time
	thisStack            []This
	packageStack         []packageDef
	output               *strings.Builder
	lastStruct           interface{}
	result               interface{}
	programCounter       int
	stackPointer         int
	framePointer         int
	line                 int
	blockDepth           int
	argCountDelta        int
	threadID             int32
	fullSymbolScope      bool
	running              bool
	Static               bool
	debugging            bool
	singleStep           bool
	breakOnReturn        bool
	stepOver             bool
	throwUncheckedErrors bool
	fullStackTrace       bool
}

// NewContext generates a new context. It must be passed a symbol table and a bytecode
// array. A context holds the runtime state of a given execution unit (program counter,
// runtime stack, symbol table) and is used to actually run bytecode. The bytecode
// can continue to be modified after it is associated with a context.
func NewContext(s *symbols.SymbolTable, b *ByteCode) *Context {
	name := ""
	if b != nil {
		name = b.Name
	}

	// Determine whether static data typing is in effect. This is
	// normally off, but can be set by a global variable (which is
	// ultimately set by a profile setting or CLI option).
	static := false
	if s, ok := s.Get("__static_data_types"); ok {
		static = util.GetBool(s)
	}

	// If we weren't given a table, create an empty temp table.
	if s == nil {
		s = symbols.NewSymbolTable("")
	}

	// Create the context object.
	ctx := Context{
		Name:                 name,
		threadID:             atomic.AddInt32(&nextThreadID, 1),
		bc:                   b,
		programCounter:       0,
		stack:                make([]interface{}, InitialStackSize),
		stackPointer:         0,
		framePointer:         0,
		running:              false,
		Static:               static,
		line:                 0,
		symbols:              s,
		fullSymbolScope:      true,
		thisStack:            nil,
		throwUncheckedErrors: persistence.GetBool(defs.ThrowUncheckedErrorsSetting),
		fullStackTrace:       persistence.GetBool(defs.FullStackTraceSetting),
		packageStack:         make([]packageDef, 0),
		tryStack:             make([]TryInfo, 0),
		rangeStack:           make([]*Range, 0),
		timerStack:           make([]time.Time, 0),
	}
	contextPointer := &ctx
	contextPointer.SetByteCode(b)

	return contextPointer
}

// GetLine retrieves the current line number from the
// original source being executed. This is stored in the
// context every time an AtLine instruction is executed.
func (c *Context) GetLine() int {
	return c.line
}

// SetDebug turns debugging mode on or off for the current
// context.
func (c *Context) SetDebug(b bool) *Context {
	c.debugging = b
	c.singleStep = true

	return c
}

// SetFullSymbolScope sets the flag that indicates if a
// symbol table read can "see" a symbol outside the current
// function. The default is off, which means symbols are not
// visible outside the function unless they are in the global
// symbol table. If true, then a symbol can be read from any
// level of the symbol table parentage chain.
func (c *Context) SetFullSymbolScope(b bool) *Context {
	c.fullSymbolScope = b

	return c
}

// SetPC sets the program counter (PC) which indicates the
// next instruction number to execute.
func (c *Context) SetPC(pc int) {
	c.programCounter = pc
}

// SetGlobal stores a value in a the global symbol table that is
// at the top of the symbol table chain.
func (c *Context) SetGlobal(name string, value interface{}) *errors.EgoError {
	return c.symbols.Root().Set(name, value)
}

// EnableConsoleOutput tells the context to begin capturing all output normally generated
// from Print and Newline into a buffer instead of going to stdout.
func (c *Context) EnableConsoleOutput(flag bool) *Context {
	ui.Debug(ui.AppLogger, ">>> Console output set to %v", flag)

	if !flag {
		var b strings.Builder
		c.output = &b
	} else {
		c.output = nil
	}

	return c
}

// GetOutput retrieves the output buffer. This is the buffer that
// contains all Print and related bytecode instruction output. This
// is used when output capture is enabled, which typically happens
// when a program is running as a Web service.
func (c *Context) GetOutput() string {
	if c.output != nil {
		return c.output.String()
	}

	return ""
}

// SetTracing turns tracing on or off for the current context. When
// tracing is on, each time an instruction is executed, the current
// instruction and the top few items on the stack are printed to
// the console.
func (c *Context) SetTracing(b bool) {
	ui.SetLogger(ui.TraceLogger, b)
}

func (c *Context) Tracing() bool {
	return ui.LoggerIsActive(ui.TraceLogger)
}

// SetTokenizer sets a tokenizer in the current context for use by
// tracing and debugging operations. This gives those functions
// access to the token stream used to compile the bytecode in this
// context.
func (c *Context) SetTokenizer(t *tokenizer.Tokenizer) *Context {
	c.tokenizer = t

	return c
}

// GetTokenizer gets the tokenizer in the current context for
// tracing and debugging.
func (c *Context) GetTokenizer() *tokenizer.Tokenizer {
	return c.tokenizer
}

// AppendSymbols appends a symbol table to the current context.
// This is used to add in compiler maps, for example.
func (c *Context) AppendSymbols(s *symbols.SymbolTable) {
	for k, v := range s.Symbols {
		_ = c.symbols.SetAlways(k, v)
	}
}

// SetByteCode attaches a new bytecode object to the current run context.
func (c *Context) SetByteCode(b *ByteCode) {
	c.bc = b
}

// SetSingleStep enables or disables single-step mode. This has no
// effect if debugging is not active.
func (c *Context) SetSingleStep(b bool) {
	c.singleStep = b
}

// SingleStep retrieves the current single-step setting for this
// context. This is used in the debugger to know how to handle
// break operations.
func (c *Context) SingleStep() bool {
	return c.singleStep
}

// SetStepOver determines if single step operations step over a
// function call, or step into it.
func (c *Context) SetStepOver(b bool) {
	c.stepOver = b
}

// GetModuleName returns the name of the current module (typically
// the function name or program name).
func (c *Context) GetModuleName() string {
	return c.bc.Name
}

// Pop removes the top-most item from the stack.
func (c *Context) Pop() (interface{}, *errors.EgoError) {
	if c.stackPointer <= 0 || len(c.stack) < c.stackPointer {
		return nil, c.newError(errors.ErrStackUnderflow)
	}

	c.stackPointer = c.stackPointer - 1
	v := c.stack[c.stackPointer]

	return v, nil
}

// FormatStack formats the stack for tracing output.
func FormatStack(syms *symbols.SymbolTable, s []interface{}, newlines bool) string {
	var b strings.Builder

	if len(s) == 0 {
		return "<empty>"
	}

	if newlines {
		b.WriteString("stack elements:\n")
	}

	for n := len(s) - 1; n >= 0; n = n - 1 {
		if n < len(s)-1 {
			b.WriteString(", ")

			if newlines {
				b.WriteString("\n")
			}
		}

		if newlines {
			b.WriteString(fmt.Sprintf("%90s      [%2d]:   ", " ", n))
		}

		// If it's a string, escape the newlines for readability.
		if stringValue, ok := s[n].(string); ok {
			stringValue = strings.ReplaceAll(stringValue, "\t", "\\t")
			stringValue = strings.ReplaceAll(stringValue, "\n", "\\n")
			b.WriteString(stringValue)
		} else {
			b.WriteString(util.Format(s[n]))
		}

		if !newlines && b.Len() > 50 {
			return b.String()[:50] + "..."
		}
	}

	return b.String()
}

// constantSet is a helper function to define a constant value.
func (c *Context) constantSet(name string, v interface{}) *errors.EgoError {
	return c.symbols.SetConstant(name, v)
}

// symbolIsConstant is a helper function to define a constant value.
func (c *Context) symbolIsConstant(name string) bool {
	return c.symbols.IsConstant(name)
}

// symbolGet is a helper function that retrieves a symbol value from the associated
// symbol table.
func (c *Context) symbolGet(name string) (interface{}, bool) {
	v, found := c.symbols.Get(name)

	return v, found
}

// symbolSet is a helper function that sets a symbol value in the associated
// symbol table.
func (c *Context) symbolSet(name string, value interface{}) *errors.EgoError {
	return c.symbols.Set(name, value)
}

// symbolSetAlways is a helper function that sets a symbol value in the associated
// symbol table.
func (c *Context) symbolSetAlways(name string, value interface{}) *errors.EgoError {
	return c.symbols.SetAlways(name, value)
}

// symbolDelete deletes a symbol from the current context.
func (c *Context) symbolDelete(name string) *errors.EgoError {
	return c.symbols.Delete(name, false)
}

// symbolCreate creates a symbol.
func (c *Context) symbolCreate(name string) *errors.EgoError {
	return c.symbols.Create(name)
}

// stackPush puts a new items on the stack.
func (c *Context) stackPush(v interface{}) *errors.EgoError {
	if c.stackPointer >= len(c.stack) {
		c.stack = append(c.stack, make([]interface{}, GrowStackBy)...)
	}

	c.stack[c.stackPointer] = v
	c.stackPointer = c.stackPointer + 1

	return nil
}

// checkType is a utility function used to determine if a given value
// could be stored in a named symbol. When the value is nil or static
// type checking is disabled (the default) then no action occurs.
//
// Otherwise, the symbol name is used to look up the current value (if
// any) of the symbol. If it exists, then the type of the value being
// proposed must match the type of the existing value.
func (c *Context) checkType(name string, value interface{}) *errors.EgoError {
	var err *errors.EgoError

	if !c.Static || value == nil {
		return err
	}

	if oldValue, ok := c.symbolGet(name); ok {
		if oldValue == nil {
			return err
		}

		if reflect.TypeOf(value) != reflect.TypeOf(oldValue) {
			err = c.newError(errors.ErrInvalidVarType)
		}
	}

	return err
}

func (c *Context) Result() interface{} {
	return c.result
}

func (c *Context) popSymbolTable() {
	name := c.symbols.Name
	c.symbols = c.symbols.Parent

	for strings.HasPrefix(c.symbols.Name, "pkg func ") {
		if c.symbols.Parent == nil {
			break
		}

		c.symbols = c.symbols.Parent
	}

	ui.Debug(ui.SymbolLogger, "(%d) pop symbol table; \"%s\" => \"%s\"",
		c.threadID, name, c.symbols.Name)
}
