// Package config implements utility routines for populating and managing configuration.
package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// ParseFromEnvironment walks a input looking for corresponding environment variables, which if found
// will update the structures value. Only exported fields of structured inputs are considered.

// Environment variable keys are determined by converting the key reference to upper snake case,
// splitting on any words or child references. eg.
//
//	FooBar.Baz -> FOO_BAR_BAZ
//
// Basic data types can also be used by populating the prefix argument to match the exact corresponding
// environment variable key name.
func LoadFromEnvironment[T any, PtrT *T](pt PtrT, prefix string) error {
	ptr := reflect.ValueOf(pt)
	err := loadRecursive(reflect.Indirect(ptr), prefix)
	if err != nil {
		return err
	}
	return nil
}

// MustLoadFromEnvironment behaves the same as [LoadFromEnvironment] but will panic instead of
// returning any errors.
func MustLoadFromEnvironment[T any, PtrT *T](pt PtrT, prefix string) {
	ptr := reflect.ValueOf(pt)
	err := loadRecursive(reflect.Indirect(ptr), prefix)
	if err != nil {
		panic(err)
	}
}

func loadRecursive(val reflect.Value, environmentKey string) error {
	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if !field.CanSet() {
				continue
			}

			fieldName := val.Type().Field(i).Name
			err := loadRecursive(field, camelToUpperSnakeCase(fieldName, environmentKey))
			if err != nil {
				return err
			}
		}
	}

	// For slices and arrays, populate sparse structures. We don't allow specifying
	// config differently per index.
	if val.Kind() == reflect.Array || val.Kind() == reflect.Slice {
		for j := 0; j < val.Len(); j++ {
			err := loadRecursive(val.Index(j), environmentKey)
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
