package datatypes

import (
	"sort"
	"strings"

	"github.com/tucats/ego/errors"
)

type EgoStruct struct {
	typeDef      Type
	static       bool
	readonly     bool
	strongTyping bool
	replica      int
	fields       map[string]interface{}
}

func NewStruct(t Type) *EgoStruct {
	// IF this is a user type, get the base type.
	baseType := t
	for baseType.IsTypeDefinition() {
		baseType = *baseType.BaseType()
	}

	// It must be a structure
	if baseType.kind != StructKind {
		return nil
	}

	// If there are fields defined, this is static.
	static := true
	if baseType.fields == nil || len(baseType.fields) == 0 {
		static = false
	}

	// Create the fields structure, and fill it in using the field
	// names and the zero-type of each kind
	fields := map[string]interface{}{}

	for k, v := range baseType.fields {
		fields[k] = InstanceOfType(v)
	}

	// Create the structure and pass it back.
	result := EgoStruct{
		typeDef: baseType,
		static:  static,
		fields:  fields,
	}

	return &result
}

func NewStructFromMap(m map[string]interface{}) *EgoStruct {
	t := Structure()

	if value, ok := GetMetadata(m, TypeMDKey); ok {
		t = GetType(value)
	} else {
		for k, v := range m {
			_ = t.DefineField(k, TypeOf(v))
		}
	}

	static := (len(m) > 0)
	if value, ok := GetMetadata(m, StaticMDKey); ok {
		static = GetBool(value)
	}

	readonly := false
	if value, ok := GetMetadata(m, ReadonlyMDKey); ok {
		readonly = GetBool(value)
	}

	result := EgoStruct{
		static:   static,
		typeDef:  t,
		readonly: readonly,
		fields:   m,
	}

	return &result
}

func (s *EgoStruct) IsReplica() *EgoStruct {
	s.replica++

	return s
}

func (s *EgoStruct) SetTyping(b bool) *EgoStruct {
	s.strongTyping = b

	return s
}

func (s *EgoStruct) SetReadonly(b bool) *EgoStruct {
	s.readonly = b

	return s
}

func (s *EgoStruct) SetStatic(b bool) *EgoStruct {
	s.static = b

	return s
}

func (s *EgoStruct) Get(name string) (interface{}, bool) {
	value, ok := s.fields[name]

	return value, ok
}

func (s EgoStruct) ToMap() map[string]interface{} {
	result := map[string]interface{}{}

	for k, v := range s.fields {
		result[k] = DeepCopy(v)
	}

	return result
}

func (s *EgoStruct) Set(name string, value interface{}) *errors.EgoError {
	if s.readonly {
		return errors.New(errors.ReadOnlyError)
	}

	// Is it a readonly symbol name and it already exists? If so, fail...
	if name[0:1] == "_" {
		_, ok := s.fields[name]
		if ok {
			return errors.New(errors.ReadOnlyError)
		}
	}

	if s.static {
		_, ok := s.fields[name]
		if !ok {
			return errors.New(errors.InvalidFieldError)
		}
	}

	if s.strongTyping && s.typeDef.fields != nil {
		if t, ok := s.typeDef.fields[name]; ok {
			if !IsType(value, t) {
				return errors.New(errors.InvalidTypeError)
			}
		}
	}

	s.fields[name] = value

	return nil
}

// Make a copy of the current structure object.
func (s EgoStruct) Copy() *EgoStruct {
	result := NewStructFromMap(s.fields)
	result.readonly = s.readonly
	result.static = s.static
	result.replica = s.replica + 1
	result.strongTyping = s.strongTyping

	return result
}

func (s EgoStruct) FieldNames() []string {
	keys := make([]string, 0)

	for k := range s.fields {
		if !strings.HasPrefix(k, "__") {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)

	return keys
}

func (s EgoStruct) FieldNamesArray() *EgoArray {
	keys := s.FieldNames()
	keyValues := make([]interface{}, len(keys))

	for i, v := range keys {
		keyValues[i] = v
	}

	return NewFromArray(StringType, keyValues)
}

func (s EgoStruct) TypeString() string {
	if s.typeDef.IsTypeDefinition() {
		return s.typeDef.name
	}

	return s.typeDef.String()
}

func (s EgoStruct) String() string {
	if len(s.fields) == 0 {
		return "{}"
	}

	keys := make([]string, 0)
	b := strings.Builder{}

	b.WriteString("{ ")

	for k := range s.fields {
		if !strings.HasPrefix(k, "__") {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)

	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}

		v := s.fields[k]

		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(Format(v))
	}

	b.WriteString(" }")

	return b.String()
}

func (s EgoStruct) Reflect() *EgoStruct {
	m := map[string]interface{}{}

	m["type"] = s.TypeString()
	m["basetype"] = "struct"
	m["members"] = s.FieldNamesArray()
	m["replicas"] = s.replica
	m["readonly"] = s.readonly
	m["static"] = s.static

	return NewStructFromMap(m)
}
