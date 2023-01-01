package bytecode

import (
	"fmt"
	"reflect"

	"github.com/tucats/ego/app-cli/ui"
	"github.com/tucats/ego/datatypes"
	"github.com/tucats/ego/defs"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/symbols"
)

type OptimizerOperation int

const (
	OptNothing OptimizerOperation = iota
	OptStore
	OptRead
	OptCount
	OptRunConstantFragment
)

type OptimizerToken struct {
	Name      string
	Operation OptimizerOperation
	Register  int
	Value     interface{}
}

type Optimization struct {
	Description string
	Debug       bool
	Source      []Instruction
	Replacement []Instruction
}

// Optimize runs a peep-hold optimizer over the bytecode.
func (b *ByteCode) Optimize() (int, *errors.EgoError) {
	startingSize := b.emitPos
	count := 0

	// Figure out the maximum pattern size, since we'll need this for backing
	// up the bytecode scanner after each patch operation.
	maxPatternSize := 0
	for _, optimization := range Optimizations {
		if max := len(optimization.Source); max > maxPatternSize {
			maxPatternSize = max
		}
	}

	// Starting at each sequential bytecode, see if any of the patterns
	// match.
	for idx := 0; idx < b.emitPos; idx++ {
		found := false

		// Scan over all the available optimizations.
		for _, optimization := range Optimizations {
			operandValues := map[string]OptimizerToken{}
			registers := make([]interface{}, 5)
			found = true

			// Is there a branch INTO this pattern? If so, then it
			// cannot be optimized.
			for _, i := range b.instructions {
				if i.Operation > BranchInstructions {
					destination := datatypes.GetInt(i.Operand)
					if destination >= idx && destination < idx+len(optimization.Source) {
						found = false

						break
					}
				}
			}

			if !found {
				continue
			}

			// Debugging trap for optimization in "main"
			if optimization.Debug && b.Name == defs.Main {
				fmt.Println("DEBUG breakpoint for " + optimization.Description)
			}

			// Search each instruction in the pattern to see if it matches
			// with the instruction stream we are positioned at.
			for sourceIdx, sourceInstruction := range optimization.Source {
				if b.emitPos <= idx+sourceIdx {
					found = false

					// This optimization can't fit because we're too near the end
					// so go on to the next optimization.
					break
				}

				i := b.instructions[idx+sourceIdx]

				if sourceInstruction.Operation != i.Operation {
					found = false

					// This optimization didn't match; go to next optimization
					break
				}

				if token, ok := sourceInstruction.Operand.(OptimizerToken); ok {
					value, inMap := operandValues[token.Name]
					if inMap {
						if value.Value == i.Operand {
							// no work to do
						} else if i.Operand != sourceInstruction.Operand {
							found = false

							continue
						}
					} else {
						switch token.Operation {
						case OptCount:
							increment := 1
							if i.Operand != nil {
								increment = datatypes.GetInt(i.Operand)
							}
							registers[token.Register] = datatypes.GetInt(registers[token.Register]) + increment

						case OptStore:
							registers[token.Register] = i.Operand
						}

						operandValues[token.Name] = OptimizerToken{Name: token.Name, Value: i.Operand}
					}
				}
			}

			// Does this optimization match?
			if found {
				if count == 0 && ui.LoggerIsActive(ui.OptimizerLogger) {
					ui.Debug(ui.OptimizerLogger, "@@@ Optimizing bytecode %s @@@", b.Name)
					ui.Debug(ui.OptimizerLogger, "    Code before optimizations:")

					oldBytecodeLoggingStatus := ui.LoggerIsActive(ui.ByteCodeLogger)

					ui.SetLogger(ui.ByteCodeLogger, true)
					b.Disasm()
					ui.SetLogger(ui.ByteCodeLogger, oldBytecodeLoggingStatus)
					ui.Debug(ui.OptimizerLogger, "")
				}

				ui.Debug(ui.OptimizerLogger, "Optimization found in %s: %s", b.Name, optimization.Description)

				// Make a copy of the replacements, with the token values from the
				// source stream inserted as appropriate.
				replacements := []Instruction{}

				for _, replacement := range optimization.Replacement {
					newInstruction := replacement

					if token, ok := replacement.Operand.(OptimizerToken); ok {
						switch token.Operation {
						case OptRunConstantFragment:
							v, _ := b.executeFragment(idx, idx+len(optimization.Source))
							newInstruction.Operand = v

						case OptRead:
							newInstruction.Operand = registers[token.Register]

						default:
							newInstruction.Operand = operandValues[token.Name].Value
						}
					}

					// Second slightly more complex case, where the replacement
					// consists of multiple tokens, any of which might be drawn
					// from the valuemap.
					if tokenArray, ok := replacement.Operand.([]interface{}); ok {
						newArray := []interface{}{}

						for _, item := range tokenArray {
							if token, ok := item.(OptimizerToken); ok {
								newArray = append(newArray, operandValues[token.Name].Value)
							} else {
								newArray = append(newArray, item)
							}
						}

						newInstruction.Operand = newArray
					}

					replacements = append(replacements, newInstruction)
				}

				b.Patch(idx, len(optimization.Source), replacements)

				// Back up the pointer and continue, since we may now be part of
				// a previous pattern.
				idx = idx - maxPatternSize
				if idx < 0 {
					idx = 0
				}

				count++
			}
		}
	}

	// Now do any additional optimizations that aren't pattern-based.
	count += b.constantStructOptimizer()

	if count > 0 && ui.LoggerIsActive(ui.OptimizerLogger) {
		ui.Debug(ui.OptimizerLogger, "Found %d optimization(s) for net change in size of %d instructions", count, startingSize-b.emitPos)
		oldBytecodeLoggingStatus := ui.LoggerIsActive(ui.ByteCodeLogger)

		ui.Debug(ui.OptimizerLogger, "    Code after  optimizations:")

		ui.SetLogger(ui.ByteCodeLogger, true)
		b.Disasm()
		ui.SetLogger(ui.ByteCodeLogger, oldBytecodeLoggingStatus)
		ui.Debug(ui.OptimizerLogger, "")
	}

	return count, nil
}

func (b *ByteCode) executeFragment(start, end int) (interface{}, *errors.EgoError) {
	fragment := New("code fragment")

	for idx := start; idx < end; idx++ {
		i := b.instructions[idx]
		fragment.Emit(i.Operation, i.Operand)
	}

	s := symbols.NewSymbolTable("fragment")
	c := NewContext(s, fragment)

	_ = c.Run()

	return c.Pop()
}

func (b *ByteCode) Patch(start, deleteSize int, insert []Instruction) {
	offset := deleteSize - len(insert)

	// Start by deleting the old instructions
	instructions := append(b.instructions[:start], insert...)
	instructions = append(instructions, b.instructions[start+deleteSize:]...)

	// Scan the instructions with destinations after the insertion and update jump offsets
	for i := 0; i < len(instructions); i++ {
		if instructions[i].Operation > BranchInstructions {
			destination := datatypes.GetInt(instructions[i].Operand)
			if destination > start {
				instructions[i].Operand = destination - offset
			}
		}
	}

	b.instructions = instructions
	b.emitPos = b.emitPos - offset
}

func (b *ByteCode) constantStructOptimizer() int {
	count := 0

	for idx := 0; idx < b.emitPos; idx++ {
		i := b.instructions[idx]

		if i.Operation != Struct {
			continue
		}

		fieldCount := datatypes.GetInt(i.Operand)

		// Bogus count, let it be caught at runtime.
		if idx-fieldCount < 0 {
			continue
		}

		areConstant := true

		for idx2 := 1; idx2 <= fieldCount*2; idx2++ {
			if b.instructions[idx-idx2].Operation != Push {
				areConstant = false

				break
			}
		}

		// If they are all constant values, we can construct an array constant
		// here one time.
		if areConstant {
			var structType *datatypes.Type

			var typeModel interface{}

			m := map[string]interface{}{}

			for idx2 := 1; idx2 <= fieldCount*2; idx2 += 2 {
				name := datatypes.GetString(b.instructions[idx-idx2].Operand)
				value := b.instructions[idx-idx2-1].Operand

				if name == datatypes.TypeMDKey {
					if t, ok := value.(*datatypes.Type); ok {
						structType = t
						typeModel = t.InstanceOf(t)
					}
				}

				m[name] = value
			}

			// Add in any fields from the type model not present
			// in the new structure we're creating. We ignore any
			// function definitions in the model, as they will be
			// found later during function invocation if needed
			// by chasing the model chain.
			if typeModel != nil {
				if realModel, ok := typeModel.(*datatypes.EgoStruct); ok {
					for _, k := range realModel.FieldNames() {
						v, _ := realModel.Get(k)

						vx := reflect.ValueOf(v)
						if vx.Kind() == reflect.Ptr {
							ts := vx.String()
							if ts == defs.ByteCodeReflectionTypeString {
								continue
							}
						}

						if _, found := m[k]; !found {
							m[k] = v
						}
					}
				}
			}

			s := datatypes.NewStructFromMap(m)
			if structType != nil {
				s.AsType(structType)
			}

			b.Patch(idx-fieldCount*2, fieldCount*2+1, []Instruction{
				{
					Operation: Push,
					Operand:   s,
				},
			})
			count++
		}
	}

	return count
}
