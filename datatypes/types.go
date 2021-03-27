package datatypes

import (
	"sort"
	"strings"
	"sync"

	"github.com/tucats/ego/errors"
)

// Define data types as abstract identifiers. These are the base
// types for all other types. For example, a pointer to an integer
// in constructed from a pointerKind type that references an IntKind
// type.
const (
	undefinedKind = iota
	intKind
	floatKind
	stringKind
	boolKind
	structKind
	errorKind
	chanKind
	mapKind
	interfaceKind // alias for "any"
	pointerKind   // Pointer to some type
	arrayKind     // Array of some type
	packageKind   // A package

	minimumNativeType // Before list of Go-native types mapped to Ego types
	waitGroupKind
	mutexKind
	maximumNativeType // After list of Go-native types

	varArgs  // pseudo type used for variable argument list items
	typeKind // something defined by a type statement
)

type Type struct {
	name      string
	kind      int
	fields    map[string]Type
	functions map[string]interface{}
	keyType   *Type
	valueType *Type
}

type Field struct {
	Name string
	Type Type
}

// Return a string containing the list of reciver functions for
// this type. If there are no functions defined, it returns an
// empty string. The results are a comma-separated list of function
// names plus "()".
func (t Type) FunctionNameList() string {
	if t.functions == nil || len(t.functions) == 0 {
		return ""
	}

	b := strings.Builder{}
	b.WriteString(",")

	keys := make([]string, 0)

	for k := range t.functions {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for i, k := range keys {
		if i > 0 {
			b.WriteString(",")
		}

		b.WriteString(k)
		b.WriteString("()")
	}

	return b.String()
}

// Produce a human-readable version of the type definition.
func (t Type) String() string {
	switch t.kind {
	case typeKind:
		return t.name + " " + t.valueType.String()

	case mapKind:
		return "map[" + t.keyType.String() + "]" + t.valueType.String()

	case pointerKind:
		return "*" + t.valueType.String()

	case arrayKind:
		return "[]" + t.valueType.String()

	case structKind:
		// If there are fields, let's include that in the type info?
		b := strings.Builder{}
		b.WriteString("struct")

		if t.fields != nil && len(t.fields) > 0 {
			b.WriteString("{")

			keys := make([]string, 0)
			for k := range t.fields {
				keys = append(keys, k)
			}

			sort.Strings(keys)

			for i, k := range keys {
				if i > 0 {
					b.WriteString(", ")
				}

				b.WriteString(k)
				b.WriteString(" ")
				b.WriteString(t.fields[k].String())
			}

			b.WriteString("}")
		}

		return b.String()

	default:
		return t.name
	}
}

// Return true if this type is a pointer to something.
func (t Type) IsPointer() bool {
	return t.kind == pointerKind
}

// Return true if this type is the same as the provided type.
func (t Type) IsType(i Type) bool {
	if t.kind != i.kind {
		return false
	}

	if t.keyType != nil && i.keyType == nil {
		return false
	}

	if t.valueType != nil && i.valueType == nil {
		return false
	}

	if t.keyType != nil && i.keyType != nil {
		if !t.keyType.IsType(*i.keyType) {
			return false
		}
	}

	if t.valueType != nil && i.valueType != nil {
		if !t.valueType.IsType(*i.valueType) {
			return false
		}
	}

	return true
}

// Returns true if the current type is an array.
func (t Type) IsArray() bool {
	return t.kind == arrayKind
}

// Returns true if the current type is a type definition created
// by code (as opposed to a base type).
func (t Type) IsTypeDefinition() bool {
	return t.kind == typeKind
}

// For a given type, add a new field of the given name and type. Returns
// an error if the current type is not a structure, or if the field already
// is defined.
func (t *Type) DefineField(name string, ofType Type) *errors.EgoError {
	if t.kind != structKind {
		return errors.New(errors.InvalidStructError)
	}

	if t.fields == nil {
		t.fields = map[string]Type{}
	} else {
		if _, found := t.fields[name]; found {
			return errors.New(errors.InvalidFieldError)
		}
	}

	t.fields[name] = ofType

	return nil
}

// Retrieve the type of a field by name. The current type must
// be a structure type, and the field name must exist.
func (t Type) Field(name string) (Type, *errors.EgoError) {
	if t.kind != structKind {
		return UndefinedType, errors.New(errors.InvalidStructError)
	}

	if t.fields == nil {
		return UndefinedType, errors.New(errors.InvalidFieldError)
	}

	ofType, found := t.fields[name]
	if !found {
		return UndefinedType, errors.New(errors.InvalidFieldError)
	}

	return ofType, nil
}

// Return true if the curren type is the undefined type.
func (t Type) IsUndefined() bool {
	return t.kind == undefinedKind
}

// Add a function definition to a type. The name of the
// function is given, and an interface value (which should
// be either the bytecode for a compiled function, or a
// function pointer for a built-in function).
func (t *Type) AddFunction(name string, v interface{}) {
	if t.functions == nil {
		t.functions = map[string]interface{}{}
	}

	t.functions[name] = v
}

// Retrieve a receiver function from the given type. Returns
// nil if there is no such function.
func (t Type) Function(name string) interface{} {
	var v interface{}

	ok := false

	if t.functions != nil {
		v, ok = t.functions[name]
	}

	if !ok && t.kind == typeKind && t.valueType != nil {
		return t.valueType.Function(name)
	}

	return v
}

// For a given type, return the type of it's base type. So for
// an array, this is the type of each array element. For a pointer,
// it is the type it points to.
func (t Type) BaseType() *Type {
	return t.valueType
}

// Return the name of the type (not the same as the
// formatted string, but usually refers to a user-defined
// type name).
func (t Type) Name() string {
	return t.name
}

// TypeOf accepts an interface of arbitrary Ego or native data type,
// and returns the associated type specification, such as datatypes.intKind
// or datatypes.stringKind.
func TypeOf(i interface{}) Type {
	switch v := i.(type) {
	case *interface{}:
		baseType := TypeOf(*v)

		return Pointer(baseType)

	case *sync.WaitGroup:
		return WaitGroupType

	case **sync.WaitGroup:
		return Pointer(WaitGroupType)

	case *sync.Mutex:
		return MutexType

	case **sync.Mutex:
		return Pointer(MutexType)

	case int:
		return IntType

	case float32, float64:
		return FloatType

	case string:
		return StringType

	case bool:
		return BoolType

	case map[string]interface{}:
		// Is it a struct with an embedded type metadata item?
		if t, ok := GetMetadata(v, TypeMDKey); ok {
			if t, ok := t.(Type); ok {
				return t
			}
		}

		// Nope, apparently just an anonymous struct
		return Type{
			name: "struct",
			kind: structKind,
		}

	case *int:
		return Pointer(IntType)

	case *float32, *float64:
		return Pointer(FloatType)

	case *string:
		return Pointer(StringType)

	case *bool:
		return Pointer(BoolType)

	case *map[string]interface{}:
		return Type{
			name: "*struct",
			kind: pointerKind,
			valueType: &Type{
				name: "struct",
				kind: structKind,
			},
		}

	case *EgoMap:
		return v.Type()

	case *Channel:
		return Pointer(ChanType)

	default:
		return InterfaceType
	}
}

// IsType accepts an arbitrary value that is either an Ego or native data
// value, and a type specification, and indicates if it is of the provided
// Ego datatype indicator.
func IsType(v interface{}, t Type) bool {
	if t.kind == interfaceKind {
		return true
	}

	return t.IsType(TypeOf(v))
}

// For a given interface pointer, unwrap the pointer and return the type it
// actually points to.
func TypeOfPointer(v interface{}) Type {
	if p, ok := v.(Type); ok {
		if p.kind != pointerKind || p.valueType == nil {
			return UndefinedType
		}

		return *p.valueType
	}

	// Is this a pointer to an actual native interface?
	p, ok := v.(*interface{})
	if !ok {
		return UndefinedType
	}

	actual := *p

	return TypeOf(actual)
}

// Determine if the given value is "nil". This an be either an actual
// nil value, or a value that represents the "nil values" for the given
// type (which are recorded as the address of the zero value).
func IsNil(v interface{}) bool {
	// Is it outright a nil value?
	if v == nil {
		return true
	}

	// Is it a nil error message?
	if err, ok := v.(*errors.EgoError); ok {
		return errors.Nil(err)
	}

	// If it's not a pointer, then it can't be nil
	addr, ok := v.(*interface{})
	if !ok {
		return false
	}

	// Compare the pointer to the known "Zero values"
	// used to initialize empty pointers.
	if addr == nil {
		return true
	} else if addr == &boolInterface {
		return true
	} else if addr == &intInterface {
		return true
	} else if addr == &stringInterface {
		return true
	} else if addr == &floatInterface {
		return true
	} else if addr == &interfaceModel {
		return true
	}

	return false
}

// Is this type associated with a native Ego type that has
// extended native function support?
func IsNativeType(kind int) bool {
	return kind > minimumNativeType && kind < maximumNativeType
}

// For a given type, return the native package that contains
// it. For example, sync.WaitGroup would return "sync".
func PackageForKind(kind int) string {
	for _, item := range TypeDeclarations {
		if item.Kind.kind == kind {
			// If this is a pointer type, skip the pointer token
			if item.Tokens[0] == "*" {
				return item.Tokens[1]
			}

			return item.Tokens[0]
		}
	}

	return ""
}
