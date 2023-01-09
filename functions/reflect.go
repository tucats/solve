package functions

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/tucats/ego/data"
	"github.com/tucats/ego/defs"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/symbols"
	"github.com/tucats/ego/util"
)

func Reflect(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	vv := reflect.ValueOf(args[0])
	ts := vv.String()

	// If it's a builtin function, it's description will match the signature. If it's a
	// match, find out it's name and return it as a builtin.
	if ts == "<func(*symbols.SymbolTable, []interface {}) (interface {}, error) Value>" {
		name := runtime.FuncForPC(reflect.ValueOf(args[0]).Pointer()).Name()
		name = strings.Replace(name, "github.com/tucats/ego/", "", 1)
		name = strings.Replace(name, "github.com/tucats/ego/runtime.", "", 1)

		declaration := data.GetBuiltinDeclaration(name)

		values := map[string]interface{}{
			data.TypeMDName:     "builtin",
			data.BasetypeMDName: "builtin " + name,
			"istype":            false,
		}

		if declaration != "" {
			values["declaration"] = declaration
		}

		return data.NewStructFromMap(values), nil
	}

	// If it's a bytecode.Bytecode pointer, use reflection to get the
	// Name field value and use that with the name. A function literal
	// will have no name.
	if vv.Kind() == reflect.Ptr {
		if ts == defs.ByteCodeReflectionTypeString {
			switch v := args[0].(type) {
			default:
				r := reflect.ValueOf(v).MethodByName("String").Call([]reflect.Value{})
				str := r[0].Interface().(string)

				name := strings.Split(str, "(")[0]
				if name == "" {
					name = defs.Anon
				}

				r = reflect.ValueOf(v).MethodByName("Declaration").Call([]reflect.Value{})
				fd, _ := r[0].Interface().(*data.FunctionDeclaration)

				return data.NewStructFromMap(map[string]interface{}{
					data.TypeMDName:     "func",
					data.BasetypeMDName: "func " + name,
					"istype":            false,
					"declaration":       makeDeclaration(fd),
				}), nil
			}
		}
	}

	if m, ok := args[0].(*data.EgoStruct); ok {
		return m.Reflect(), nil
	}

	if m, ok := args[0].(*data.Type); ok {
		return m.Reflect(), nil
	}

	// Is it an Ego package?
	if m, ok := args[0].(*data.EgoPackage); ok {
		// Make a list of the visible member names
		memberList := []string{}

		for _, k := range m.Keys() {
			if !strings.HasPrefix(k, data.MetadataPrefix) {
				memberList = append(memberList, k)
			}
		}

		// Sort the member list and forge it into an Ego array
		members := util.MakeSortedArray(memberList)

		result := map[string]interface{}{}
		result[data.MembersMDName] = members
		result[data.TypeMDName] = "*package"
		result["native"] = false
		result["istype"] = false
		result["imports"] = m.Imported()
		result["builtins"] = m.Builtins()

		t := data.TypeOf(m)
		if t.IsTypeDefinition() {
			result[data.TypeMDName] = t.Name()
			result[data.BasetypeMDName] = data.PackageTypeName
		}

		return data.NewStructFromMap(result), nil
	}

	// Is it an pionter to an Ego package?
	if m, ok := args[0].(*data.EgoPackage); ok {
		// Make a list of the visible member names
		memberList := []string{}

		for _, k := range m.Keys() {
			if !strings.HasPrefix(k, data.MetadataPrefix) {
				memberList = append(memberList, k)
			}
		}

		// Sort the member list and forge it into an Ego array
		members := util.MakeSortedArray(memberList)

		result := map[string]interface{}{}
		result[data.MembersMDName] = members
		result[data.TypeMDName] = "*package"
		result["native"] = false
		result["istype"] = false

		t := data.TypeOf(m)
		if t.IsTypeDefinition() {
			result[data.TypeMDName] = t.Name()
			result[data.BasetypeMDName] = data.PackageTypeName
		}

		return data.NewStructFromMap(result), nil
	}

	// Is it an Ego array datatype?
	if m, ok := args[0].(*data.EgoArray); ok {
		// What is the name of the base type value? This will always
		// be an array of interface{} unless this is []byte in which
		// case the native type is []byte as well.
		btName := "[]interface{}"
		if m.ValueType().Kind() == data.ByteType.Kind() {
			btName = "[]byte"
		}

		// Make a list of the visible member names
		result := map[string]interface{}{
			data.SizeMDName:     m.Len(),
			data.TypeMDName:     m.TypeString(),
			data.BasetypeMDName: btName,
			"istype":            false,
		}

		return data.NewStructFromMap(result), nil
	}

	if e, ok := args[0].(errors.EgoErrorMsg); ok {
		wrappedError := e.Unwrap()

		if e.Is(errors.ErrUserDefined) {
			text := data.String(e.GetContext())

			return data.NewStructFromMap(map[string]interface{}{
				data.TypeMDName:     "error",
				data.BasetypeMDName: "error",
				"error":             wrappedError.Error(),
				"text":              text,
				"istype":            false,
			}), nil
		}

		return data.NewStructFromMap(map[string]interface{}{
			data.TypeMDName:     "error",
			data.BasetypeMDName: "error",
			"error":             wrappedError.Error(),
			"text":              e.Error(),
			"istype":            false,
		}), nil
	}

	typeString, err := Type(s, args)
	if err == nil {
		result := map[string]interface{}{
			data.TypeMDName:     typeString,
			data.BasetypeMDName: typeString,
			"istype":            false,
		}

		return data.NewStructFromMap(result), nil
	}

	return nil, err
}

// Type implements the type() function.
func Type(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	switch v := args[0].(type) {
	case *data.EgoMap:
		return v.TypeString(), nil

	case *data.EgoArray:
		return v.TypeString(), nil

	case *data.EgoStruct:
		return v.TypeString(), nil

	case data.EgoStruct:
		return v.TypeString(), nil

	case nil:
		return "nil", nil

	case error:
		return "error", nil

	case *data.Channel:
		return "chan", nil

	case *data.Type:
		typeName := v.String()

		space := strings.Index(typeName, " ")
		if space > 0 {
			typeName = typeName[space+1:]
		}

		return "type " + typeName, nil

	case *data.EgoPackage:
		t := data.TypeOf(v)

		if t.IsTypeDefinition() {
			return t.Name(), nil
		}

		return t.String(), nil

	case *interface{}:
		tt := data.TypeOfPointer(v)

		return "*" + tt.String(), nil

	case func(s *symbols.SymbolTable, args []interface{}) (interface{}, error):
		return "<builtin>", nil

	default:
		tt := data.TypeOf(v)
		if tt.IsUndefined() {
			vv := reflect.ValueOf(v)
			if vv.Kind() == reflect.Func {
				return "builtin", nil
			}

			if vv.Kind() == reflect.Ptr {
				ts := vv.String()
				if ts == defs.ByteCodeReflectionTypeString {
					return "func", nil
				}

				return fmt.Sprintf("ptr %s", ts), nil
			}

			return "unknown", nil
		}

		vv := reflect.ValueOf(v)
		if vv.Kind() == reflect.Ptr {
			ts := vv.String()
			if ts == defs.ByteCodeReflectionTypeString {
				return "func", nil
			}

			return fmt.Sprintf("ptr %s", ts), nil
		}

		return tt.String(), nil
	}
}

// SizeOf returns the size in bytes of an arbibrary object.
func SizeOf(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	size := data.RealSizeOf(args[0])

	return size, nil
}

// makeDeclaration constructs a native data structure describing a function declaration.
func makeDeclaration(fd *data.FunctionDeclaration) *data.EgoStruct {
	parameterType := data.TypeDefinition(data.NoName, &data.StructType)
	parameterType.DefineField("name", &data.StringType)
	parameterType.DefineField(data.TypeMDName, &data.StringType)

	parameters := data.NewArray(parameterType, len(fd.Parameters))

	for n, i := range fd.Parameters {
		parameter := data.NewStruct(parameterType)
		_ = parameter.Set("name", i.Name)
		_ = parameter.Set(data.TypeMDName, i.ParmType.Name())

		_ = parameters.Set(n, parameter)
	}

	returnTypes := make([]interface{}, len(fd.ReturnTypes))

	for i, t := range fd.ReturnTypes {
		returnTypes[i] = t.TypeString()
	}

	declaration := make(map[string]interface{})

	declaration["name"] = fd.Name
	declaration["parameters"] = parameters
	declaration["returns"] = data.NewArrayFromArray(&data.StringType, returnTypes)

	return data.NewStructFromMap(declaration)
}
