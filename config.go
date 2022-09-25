package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// ParseFromEnvironment walks a *struct and populates its values from environment variables
// matching its keys converted to Upper Snake Case. Nested struct values are treated the same as
// repated words. That is foo.bar will look for environment variable FOO_BAR as will FooBar. Slices
// look like nested structs, you cannot set values for specific elements.
// Fields that do not find matching environment variables are preserved and not validated in any way.
// Unexported fields are not parsed and struct keys are expected to be in camel case.
// FIXME: is there a better form than any/interface{} to only allow pointers so structs here?
func ParseFromEnvironment[T any, PtrT *T](pt PtrT) error {
	ptr := reflect.ValueOf(pt)
	err := parseRecursive(reflect.Indirect(ptr), "")
	if err != nil {
		return err
	}
	return nil
}

// parseRecursive recursively parses an aritrary structure building up an expected environment
// variable path to find corresponding configuration values at.
func parseRecursive(val reflect.Value, environmentKey string) error {
	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if !field.CanSet() {
				continue
			}

			fieldName := val.Type().Field(i).Name
			err := parseRecursive(field, camelToUpperSnakeCase(fieldName, environmentKey))
			if err != nil {
				return err
			}
		}
	}

	// For slices and arrays, populate sparse structures. We don't allow specifying
	// config differently per index.
	if val.Kind() == reflect.Array || val.Kind() == reflect.Slice {
		for j := 0; j < val.Len(); j++ {
			err := parseRecursive(val.Index(j), environmentKey)
			if err != nil {
				return err
			}
		}
	}

	// leaf nodes
	if environmentValue, ok := os.LookupEnv(environmentKey); ok {
		switch val.Kind() {
		case reflect.String:
			val.SetString(environmentValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.Atoi(environmentValue)
			if err != nil {
				return fmt.Errorf("failed to parse environment key: %s to int: %w", environmentKey, err)
			}
			val.SetInt(int64(n))
		case reflect.Bool:
			b, err := strconv.ParseBool(environmentValue)
			if err != nil {
				return fmt.Errorf("failed to parse environment key: %s to bool: %w", environmentKey, err)
			}
			val.SetBool(b)
		default:
			return fmt.Errorf("failed to parse key: %s, unsupported field type: %s", environmentKey, val.Kind())
		}
	}
	return nil
}

// camelToUpperSnakeCase converts a string in camel case to upper snake case
// with an optional prefix prepended.
func camelToUpperSnakeCase(in string, prefix string) string {
	p := strings.ToUpper(prefix)

	if in == "" {
		return p
	}

	var s string
	var upperSequence bool
	for pos, char := range in {
		isUpper := unicode.IsUpper(char)
		if pos != 0 && isUpper && !upperSequence {
			s += "_"
		}
		if isUpper {
			s += string(char)
			upperSequence = true
			continue
		}
		s += string(unicode.ToUpper(char))
		upperSequence = false
	}

	if p == "" {
		return s
	}

	return p + "_" + s
}
