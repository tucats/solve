package datatypes

import (
	"strings"

	"github.com/tucats/gopackages/util"
)

// For any given _Ego_ object type, remove any metadata from it
// and return a sanitized copy. This is used to send HTTP response
// values, for example.
func Sanitize(v interface{}) interface{} {
	switch v := v.(type) {
	case *EgoArray:
		return v.data

	case *EgoMap:
		result := map[string]interface{}{}
		keys := v.Keys()

		for _, key := range keys {
			value, _, _ := v.Get(key)
			result[util.GetString(key)] = value
		}

		return result

	case map[string]interface{}:
		result := map[string]interface{}{}

		for key, value := range v {
			if !strings.HasPrefix(key, "__") {
				result[key] = value
			}
		}

		return result

	// For anything else, just return the thing we were given.
	default:
		return v
	}
}