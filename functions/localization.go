package functions

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/tucats/ego/datatypes"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/symbols"
)

func I18nLanguage(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	if len(args) > 0 {
		return nil, errors.EgoError(errors.ErrWrongParameterCount)
	}

	language := os.Getenv("LANG")

	if pos := strings.Index(language, "_"); pos > 0 {
		language = language[:pos]
	}

	if language == "" {
		language = "en"
	}

	return language, nil
}

func I18nT(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	parameters := map[string]string{}
	property := datatypes.GetString(args[0])

	language := os.Getenv("LANG")

	if pos := strings.Index(language, "_"); pos > 0 {
		language = language[:pos]
	}

	if len(args) > 1 {
		value := args[1]
		if egoMap, ok := value.(*datatypes.EgoMap); ok {
			keys := egoMap.Keys()
			for _, key := range keys {
				value, _, _ := egoMap.Get(key)
				parameters[datatypes.GetString(key)] = datatypes.GetString(value)
			}
		} else if egoStruct, ok := value.(*datatypes.EgoStruct); ok {
			fields := egoStruct.FieldNames()
			for _, field := range fields {
				value := egoStruct.GetAlways(field)
				parameters[field] = datatypes.GetString(value)
			}
		} else if value != nil {
			return nil, errors.EgoError(errors.ErrArgumentType)
		}
	}

	if len(args) > 2 {
		language = datatypes.GetString(args[2])
	}

	if language == "" {
		language = "en"
	}

	language = strings.ToLower(language)

	// Find the localization data
	localizedMap, found := s.Get("__localization")
	if !found {
		return property, nil
	}

	// Find the language
	if languages, ok := localizedMap.(*datatypes.EgoStruct); ok {
		stringMap, found := languages.Get(language)
		if !found {
			// If not found, assume english
			stringMap, found = languages.Get("en")
			if !found {
				return property, nil
			}
		}

		if localizedStrings, ok := stringMap.(*datatypes.EgoStruct); ok {
			message, found := localizedStrings.Get(property)
			if !found {
				return property, nil
			}

			msgString := datatypes.GetString(message)
			t := template.New(property)

			t, e := t.Parse(msgString)
			if e != nil {
				return nil, errors.EgoError(e)
			}

			var r bytes.Buffer

			e = t.Execute(&r, parameters)
			if e != nil {
				return nil, errors.EgoError(e)
			}

			return r.String(), nil
		}
	}

	return property, nil
}
