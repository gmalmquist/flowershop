// Functions for conveniently dealing with arbitrary JSON objects.
//
// The main interface here is `JsonValue`, which wraps a `[]byte` representation of an encoded
// JSON object.
//
// `JsonValue` then has methods to introspect on the underlying object and convert it to
// different native golang types (where possible). This is done in a permissive and ergonomic
// way; e.g., the `String()` method will return a String representation of any legal JSON value,
// even if the underlying type was not originally a String.
package jsv

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Wraps a `[]byte` representation of an encoded JSON value.
type JsonValue []byte

// A validation function (see the `github.com/gmalmquist/flowershop/jsv/validate` package).
type Validate func(JsonValue) error

var falsies map[string]bool = func() map[string]bool {
	arr := []string{"false", "", "[]", "{}", "0", "''", `""`, "null"}
	m := map[string]bool{}
	for _, x := range arr {
		m[x] = true
	}
	return m
}()

// Marshals this value to json using the standard library `encoding/json`, and wraps it in a
// `JsonValue` type.
func Marshal(ref any) (JsonValue, error) {
	bytes, err := json.Marshal(ref)
	if err != nil {
		return nil, err
	}
	return JsonValue(bytes), err
}

// Attempts to unmarshal the given JsonValue into the provided reference. An error will be
// returned if the types don't match.
func (v JsonValue) Unmarshal(ref any) error {
	return json.Unmarshal(v, ref)
}

// Returns a new map by marshalling all the values in the input map.
func MarshalMap(m map[string]any) (map[string]JsonValue, error) {
	blob, err := Marshal(m)
	if err != nil {
		return nil, err
	}
	return blob.Map()
}

func (v JsonValue) nota(kind string) error {
	return fmt.Errorf(
		"%v is not a %v", v.JSON(), kind,
	)
}

// Returns a `string` representation of this JSON object.
//
// If the underlying type is an actual string to begin with, this method returns the value
// unchanged. If it's anything else, it returns the string encoded JSON representation.
//
// Leading and trailing whitespace is trimmed.
func (v JsonValue) String() string {
	var parsed string
	if err := v.Unmarshal(&parsed); err != nil {
		return v.JSON()
	}
	return strings.TrimSpace(parsed)
}

// Returns the raw JSON representation of this object encoded in a string.
func (v JsonValue) JSON() string {
	return string(v)
}

// Returns true if the JSON object looks falsey.
func (v JsonValue) Falsey() bool {
	return falsies[v.String()]
}

// Returns true if the JSON object looks truthy.
func (v JsonValue) Truthy() bool {
	return !falsies[v.String()]
}

// Attempts to convert the underlying JSON value to an integer.
func (v JsonValue) Integer() (int64, error) {
	var parsed int64
	err := v.Unmarshal(&parsed)
	if err == nil {
		return parsed, nil
	}
	parsed, err = strconv.ParseInt(v.String(), 10, 64)
	if err != nil {
		return parsed, v.nota("integer")
	}
	return parsed, err
}

// Attempts to convert the underlying JSON value to a float.
func (v JsonValue) Float() (float64, error) {
	var parsed float64
	err := v.Unmarshal(&parsed)
	if err == nil {
		return parsed, nil
	}
	parsed, err = strconv.ParseFloat(v.String(), 10)
	if err != nil {
		return parsed, v.nota("integer")
	}
	return parsed, err
}

// Attempts to convert the underying JSON value to a string array.
func (v JsonValue) StringArray() ([]string, error) {
	var parsed []string
	if err := v.Unmarshal(&parsed); err != nil {
		return parsed, v.nota("string array")
	}
	return parsed, nil
}

// Attempts to convert the underying JSON value to an array of other `JsonValue`s.
func (v JsonValue) Array() ([]JsonValue, error) {
	var results []JsonValue
	var parsed []any
	if err := v.Unmarshal(&parsed); err != nil {
		return results, v.nota("array")
	}
	for _, a := range parsed {
		m, err := json.Marshal(a)
		if err != nil {
			return results, err
		}
		results = append(results, m)
	}
	return results, nil
}

// If the underlying value is a map, returns it.
//
// Otherwise, returns an empty map.
func (v JsonValue) MapOrEmpty() map[string]JsonValue {
	m, err := v.Map()
	if err != nil {
		return map[string]JsonValue{}
	}
	return m
}

// If the underlying JSON object is a map, returns that map as a map of `string` -> `JsonValue`s.
func (v JsonValue) Map() (map[string]JsonValue, error) {
	results := map[string]JsonValue{}
	var parsed map[string]any
	if err := v.Unmarshal(&parsed); err != nil {
		return results, v.nota("map")
	}
	for key, a := range parsed {
		m, err := json.Marshal(a)
		if err != nil {
			return results, err
		}
		results[key] = m
	}
	return results, nil
}

// Creates a new validation function by AND-ing this one with the argument.
func (a Validate) And(b Validate) Validate {
	return func(value JsonValue) error {
		aerr := a(value)
		berr := b(value)
		if aerr == nil && berr == nil {
			return nil
		}
		if aerr == nil {
			return berr
		}
		if berr == nil {
			return aerr
		}
		return fmt.Errorf(
			"%v, %v", aerr, berr,
		)
	}
}

// Creates a new validation function by OR-ing this one with the argument.
func (a Validate) Or(b Validate) Validate {
	return func(value JsonValue) error {
		aerr := a(value)
		berr := b(value)
		if aerr == nil || berr == nil {
			return nil
		}
		return fmt.Errorf(
			"%v, %v", aerr, berr,
		)
	}
}

// Returns true if this value is empty (applies to strings, arrays, and objects).
//
// Numeric values are never empty. If you want a function that also returns true for zero numbers,
// you want `IsFalsey`.
func (v JsonValue) IsBlank() bool {
	txt := strings.TrimSpace(string(v))
	return txt == "" || txt == "[]" || txt == "{}" || txt == "\"\""
}

// Resolves a jq-esque `path` value.
//
// This only works on nested objects, not arrays.
func (v JsonValue) Resolve(path string) (JsonValue, error) {
	node := v
	parts := strings.Split(path, ".")
	sofar := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		m, err := node.Map()
		if err != nil {
			return nil, fmt.Errorf("%v is not a map", sofar)
		}
		var ok bool
		node, ok = m[part]
		if !ok {
			return nil, fmt.Errorf("%v.%v is nil", sofar, part)
		}
		sofar = fmt.Sprintf("%v.%v", sofar, part)
	}
	return node, nil
}

// Resolves a jq-esque `path` value, or the given default value if it is missing or blank.
//
// This only works on nested objects, not arrays.
func (v JsonValue) ResolveOr(path string, defaultValue JsonValue) JsonValue {
	r, e := v.Resolve(path)
	if e != nil || r.IsBlank() {
		return defaultValue
	}
	return r
}

// Returns this value if it is not blank, otherwise the provided default value.
//
// See `JsonValue.NotBlank`.
func (v JsonValue) NotBlankOr(defaultValue JsonValue) JsonValue {
	if v.IsBlank() {
		return defaultValue
	}
	return v
}
