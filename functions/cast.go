package functions

import (
	"math"
	"reflect"
	"strings"
	"sync"

	"github.com/tucats/ego/datatypes"
	"github.com/tucats/ego/defs"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/symbols"
)

// For a new() on an object, we won't recursively copy objects
// nested more deeply than this. Setting this too small will
// prevent complex structures from copying correctly. Too large,
// and memory could be swallowed whole.
const MaxDeepCopyDepth = 100

// Normalize coerces a value to match the type of a model value.
func Normalize(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	v1, v2 := datatypes.Normalize(args[0], args[1])

	return MultiValueReturn{Value: []interface{}{v1, v2}}, nil
}

// New implements the new() function. If an integer type number
// or a string type name is given, the "zero value" for that type
// is returned. For an array, struct, or map, a recursive copy is
// done of the members to a new object which is returned.
func New(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	// Is the type an integer? If so it's a type
	if typeValue, ok := args[0].(int); ok {
		switch reflect.Kind(typeValue) {
		case reflect.Uint8, reflect.Int8:
			return byte(0), nil

		case reflect.Int32:
			return int32(0), nil

		case reflect.Int, reflect.Int64:
			return 0, nil

		case reflect.String:
			return "", nil

		case reflect.Bool:
			return false, nil

		case reflect.Float32:
			return float32(0), nil

		case reflect.Float64:
			return float64(0), nil

		default:
			return nil, errors.New(errors.ErrInvalidType).In("new()").Context(typeValue)
		}
	}

	// Is it an actual type?
	if typeValue, ok := args[0].(*datatypes.Type); ok {
		return typeValue.InstanceOf(typeValue), nil
	}

	if typeValue, ok := args[0].(*datatypes.Type); ok {
		return typeValue.InstanceOf(typeValue), nil
	}

	// Is the type an string? If so it's a type name
	if typeValue, ok := args[0].(string); ok {
		switch strings.ToLower(typeValue) {
		case "byte":
			return byte(0), nil

		case datatypes.Int32TypeName:
			return int32(0), nil

		case "int":
			return 0, nil

		case "string":
			return "", nil

		case "bool":
			return false, nil

		case "float32":
			return float32(0), nil

		case "float64":
			return float64(0), nil

		default:
			return nil, errors.New(errors.ErrInvalidType).In("new()").Context(typeValue)
		}
	}

	// If it's a WaitGroup, make a new one. Note, have to use the switch statement
	// form here to prevent Go from complaining that the interface{} is being copied.
	// In reality, we don't care as we don't actually make a copy anyway but instead
	// make a new waitgroup object.
	switch args[0].(type) {
	case sync.WaitGroup:
		return datatypes.InstanceOfType(&datatypes.WaitGroupType), nil
	}

	// If it's a Mutex, make a new one. We hae to do this as a swtich on the type, since a
	// cast attempt will yield a warning on invalid mutex copy operation.
	switch args[0].(type) {
	case sync.Mutex:
		return datatypes.InstanceOfType(&datatypes.MutexType), nil
	}

	// If it's a channel, just return the value
	if typeValue, ok := args[0].(*datatypes.Channel); ok {
		return typeValue, nil
	}

	// If it's a native struct, it has it's own deep copy.
	if structValue, ok := args[0].(*datatypes.EgoStruct); ok {
		return datatypes.DeepCopy(structValue), nil
	}

	// @tomcole should we also handle maps and arrays here?

	// Otherwise, make a deep copy of the item.
	r := DeepCopy(args[0], MaxDeepCopyDepth)

	// If there was a user-defined type in the source, make the clone point back to it
	switch v := r.(type) {
	case nil:
		return nil, errors.New(errors.ErrInvalidValue).In("new()").Context(nil)

	case symbols.SymbolTable:
		return nil, errors.New(errors.ErrInvalidValue).In("new()").Context("symbol table")

	case func(*symbols.SymbolTable, []interface{}) (interface{}, error):
		return nil, errors.New(errors.ErrInvalidValue).In("new()").Context("builtin function")

	// No action for this group
	case byte, int32, int, int64, string, float32, float64:

	case datatypes.EgoPackage:
		// Create the replica count if needed, and update it.
		replica := 0

		if replicaX, ok := datatypes.GetMetadata(v, datatypes.ReplicaMDKey); ok {
			replica = datatypes.GetInt(replicaX) + 1
		}

		datatypes.SetMetadata(v, datatypes.ReplicaMDKey, replica)

		dropList := []string{}

		// Organize the new item by removing things that are handled via the parent.
		keys := v.Keys()
		for _, k := range keys {
			vv, _ := v.Get(k)
			// IF it's an internal function, we don't want to copy it; it can be found via the
			// __parent link to the type
			vx := reflect.ValueOf(vv)

			if vx.Kind() == reflect.Ptr {
				ts := vx.String()
				if ts == defs.ByteCodeReflectionTypeString {
					dropList = append(dropList, k)
				}
			} else {
				if vx.Kind() == reflect.Func {
					dropList = append(dropList, k)
				}
			}
		}

		for _, name := range dropList {
			v.Delete(name)
		}

	default:
		return nil, errors.New(errors.ErrInvalidType).In("new()").Context(v)
	}

	return r, nil
}

// DeepCopy makes a deep copy of an Ego data type. It should be called with the
// maximum nesting depth permitted (i.e. array index->array->array...). Because
// it calls itself recursively, this is used to determine when to give up and
// stop traversing nested data. The default is MaxDeepCopyDepth.
func DeepCopy(source interface{}, depth int) interface{} {
	if depth < 0 {
		return nil
	}

	switch v := source.(type) {
	case bool:
		return v

	case byte:
		return v

	case int32:
		return v

	case int:
		return v

	case int64:
		return v

	case string:
		return v

	case float32:
		return v

	case float64:
		return v

	case []interface{}:
		r := make([]interface{}, 0)

		for _, d := range v {
			r = append(r, DeepCopy(d, depth-1))
		}

		return r

	case *datatypes.EgoStruct:
		return v.Copy()

	case *datatypes.EgoArray:
		r := datatypes.NewArray(v.ValueType(), v.Len())

		for i := 0; i < v.Len(); i++ {
			vv, _ := v.Get(i)
			vv = DeepCopy(vv, depth-1)
			_ = v.Set(i, vv)
		}

		return r

	case *datatypes.EgoMap:
		r := datatypes.NewMap(v.KeyType(), v.ValueType())

		for _, k := range v.Keys() {
			d, _, _ := v.Get(k)
			_, _ = r.Set(k, DeepCopy(d, depth-1))
		}

		return r

	case datatypes.EgoPackage:
		r := datatypes.EgoPackage{}
		keys := v.Keys()

		for _, k := range keys {
			d, _ := v.Get(k)
			r.Set(k, DeepCopy(d, depth-1))
		}

		return r

	default:
		return v
	}
}

// Compiler-generate casting; generally always array types. This is used to
// convert numeric arrays to a different kind of array, to convert a string
// to an array of integer (rune) values, etc.
func InternalCast(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	// Target kind is the last parameter
	kind := datatypes.GetType(args[len(args)-1])
	//	if !kind.IsArray() {
	//		return nil, errors.New(errors.ErrInvalidType)
	//	}

	source := args[0]
	if len(args) > 2 {
		source = datatypes.NewArrayFromArray(&datatypes.InterfaceType, args[:len(args)-1])
	}

	if kind.IsKind(datatypes.StringKind) {
		r := strings.Builder{}

		// If the source is an array of integers, treat them as runes to re-assemble.
		if actual, ok := source.(*datatypes.EgoArray); ok && actual != nil && actual.ValueType().IsIntegerType() {
			for i := 0; i < actual.Len(); i++ {
				ch, _ := actual.Get(i)
				r.WriteRune(rune(datatypes.GetInt(ch) & math.MaxInt32))
			}
		} else {
			str := datatypes.FormatUnquoted(source)
			r.WriteString(str)
		}

		return r.String(), nil
	}

	switch actual := source.(type) {
	// Conversion of one array type to another
	case *datatypes.EgoArray:
		if kind.IsType(actual.ValueType()) {
			return actual, nil
		}

		if kind.IsKind(datatypes.StringKind) &&
			(actual.ValueType().IsIntegerType() || actual.ValueType().IsInterface()) {
			r := strings.Builder{}

			for i := 0; i < actual.Len(); i++ {
				ch, _ := actual.Get(i)
				r.WriteRune(datatypes.GetInt32(ch) & math.MaxInt32)
			}

			return r.String(), nil
		}

		elementKind := *kind.BaseType()
		r := datatypes.NewArray(kind.BaseType(), actual.Len())

		for i := 0; i < actual.Len(); i++ {
			v, _ := actual.Get(i)

			switch elementKind.Kind() {
			case datatypes.BoolKind:
				_ = r.Set(i, datatypes.GetBool(v))

			case datatypes.ByteKind:
				_ = r.Set(i, datatypes.GetByte(v))

			case datatypes.Int32Kind:
				_ = r.Set(i, datatypes.GetInt32(v))

			case datatypes.IntKind:
				_ = r.Set(i, datatypes.GetInt(v))

			case datatypes.Int64Kind:
				_ = r.Set(i, datatypes.GetInt64(v))

			case datatypes.Float32Kind:
				_ = r.Set(i, datatypes.GetFloat32(v))

			case datatypes.Float64Kind:
				_ = r.Set(i, datatypes.GetFloat64(v))

			case datatypes.StringKind:
				_ = r.Set(i, datatypes.GetString(v))

			default:
				return nil, errors.New(errors.ErrInvalidType)
			}
		}

		return r, nil

	case string:
		if kind.IsType(datatypes.Array(&datatypes.IntType)) {
			r := datatypes.NewArray(&datatypes.IntType, 0)

			for _, rune := range actual {
				r.Append(int(rune))
			}

			return r, nil
		}

		return datatypes.Coerce(source, datatypes.InstanceOfType(kind)), nil

	default:
		if kind.IsArray() {
			r := datatypes.NewArray(kind.BaseType(), 1)
			value := datatypes.Coerce(source, datatypes.InstanceOfType(kind.BaseType()))
			_ = r.Set(0, value)

			return r, nil
		}

		v := datatypes.Coerce(source, datatypes.InstanceOfType(kind))
		if v != nil {
			return datatypes.Coerce(source, datatypes.InstanceOfType(kind)), nil
		}

		return nil, errors.New(errors.ErrInvalidType)
	}
}
