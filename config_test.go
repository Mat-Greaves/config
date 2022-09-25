package config

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

func Test_LoadFromEnvironment(t *testing.T) {
	t.Run("struct input with default values", func(t *testing.T) {
		d := "some default value"
		config := struct {
			foo string
		}{
			foo: d,
		}
		err := LoadFromEnvironment(&config, "")
		if err != nil {
			t.Error(err)
		}
		if config.foo != d {
			t.Errorf("got: %s, want: %s", config.foo, d)
		}
	})

	t.Run("struct input has int field", func(t *testing.T) {
		config := struct{ Foo int }{}
		val := 1
		os.Setenv("FOO", strconv.Itoa(val))
		defer os.Clearenv()
		err := LoadFromEnvironment(&config, "")
		if err != nil {
			t.Error(err)
		}
		if config.Foo != val {
			t.Errorf("got: %d, want: %d", config.Foo, val)
		}
	})

	t.Run("struct input invalid int field", func(t *testing.T) {
		config := struct{ Foo int }{}
		val := "abc123"
		os.Setenv("FOO", val)
		defer os.Clearenv()
		err := LoadFromEnvironment(&config, "")
		if err == nil {
			t.Errorf("got: %s, expected: error", err)
		}
		if !strings.Contains(err.Error(), "failed to parse environment key: FOO to int") {
			t.Errorf("error message did not match expected format: %s", err)
		}
	})

	t.Run("struct input has bool", func(t *testing.T) {
		config := struct{ Foo bool }{}
		val := true
		os.Setenv("FOO", strconv.FormatBool(val))
		defer os.Clearenv()
		err := LoadFromEnvironment(&config, "")
		if err != nil {
			t.Error(err)
		}
		if config.Foo != val {
			t.Errorf("got: %v, want: %v", config.Foo, val)
		}
	})

	t.Run("struct input has invalid bool", func(t *testing.T) {
		config := struct{ Foo bool }{}
		os.Setenv("FOO", "trueish")
		defer os.Clearenv()
		err := LoadFromEnvironment(&config, "")
		if err == nil {
			t.Error("expected error, got nil")
		}
		if strings.Contains(err.Error(), "failed to farse environment key: Foo to bool") {
			t.Errorf("error message idd not match expected format: %s", err)
		}
	})

	t.Run("struct input with string field", func(t *testing.T) {
		config := struct{ Foo string }{}
		val := "bar"
		os.Setenv("FOO", val)
		defer os.Clearenv()
		err := LoadFromEnvironment(&config, "")
		if err != nil {
			t.Error(err)
		}
		if config.Foo != val {
			t.Errorf("got: %s, want: %s", config.Foo, val)
		}
	})

	t.Run("basic input with prefix", func(t *testing.T) {
		var foo string
		val := "bar"
		os.Setenv("FOO", val)
		defer os.Clearenv()
		err := LoadFromEnvironment(&foo, "FOO")
		if err != nil {
			t.Error(err)
		}
		if foo != val {
			t.Errorf("got: %s, want: %s", foo, val)
		}
	})

	t.Run("unsupported field type", func(t *testing.T) {
		config := struct{ Foo chan (string) }{}
		val := "bar"
		os.Setenv("FOO", val)
		err := LoadFromEnvironment(&config, "")
		if err == nil {
			t.Errorf("got: %s, want: error", err)
		}
	})

	t.Run("multi-word key name", func(t *testing.T) {
		config := struct{ FooBar string }{}
		val := "baz"
		os.Setenv("FOO_BAR", val)
		defer os.Clearenv()
		err := LoadFromEnvironment(&config, "")
		if err != nil {
			t.Error(err)
		}
		if config.FooBar != val {
			t.Errorf("got: %s, want: %s", config.FooBar, val)
		}
	})

	t.Run("nested struct", func(t *testing.T) {
		config := struct {
			Foo struct {
				Bar string
			}
		}{}
		val := "baz"
		os.Setenv("FOO_BAR", val)
		defer os.Clearenv()
		err := LoadFromEnvironment(&config, "")
		if err != nil {
			t.Error(err)
		}
		if config.Foo.Bar != val {
			t.Errorf("got: %s, want: %s", config.Foo.Bar, val)
		}
	})

	t.Run("default nested value", func(t *testing.T) {
		d := "qux"
		config := struct {
			Foo struct {
				Bar string
			}
		}{
			Foo: struct{ Bar string }{
				Bar: d,
			},
		}
		err := LoadFromEnvironment(&config, "")
		if err != nil {
			t.Error(err)
		}
		if config.Foo.Bar != d {
			t.Errorf("got: %s, want: %s", config.Foo.Bar, d)
		}
	})

	t.Run("error in nested struct", func(t *testing.T) {
		val := "abc123"
		config := struct {
			Foo struct {
				Bar int
			}
		}{
			Foo: struct{ Bar int }{},
		}
		os.Setenv("FOO_BAR", val)
		defer os.Clearenv()
		err := LoadFromEnvironment(&config, "")
		if err == nil {
			t.Errorf("got: %s, expected: error", err)
		}
		if !strings.Contains(err.Error(), "failed to parse environment key: FOO_BAR to int") {
			t.Errorf("error message did not match expected format: %s", err)
		}
	})

	t.Run("nested slice of struct", func(t *testing.T) {
		val := "qux"
		config := struct {
			Foo []struct {
				Bar string
			}
		}{
			Foo: []struct{ Bar string }{
				{
					Bar: "baz",
				},
			},
		}
		os.Setenv("FOO_BAR", val)
		defer os.Clearenv()
		err := LoadFromEnvironment(&config, "")
		if err != nil {
			t.Error(err)
		}
		for _, v := range config.Foo {
			if v.Bar != val {
				t.Errorf("got: %s, want: %s", v.Bar, val)
			}
		}
	})
}

func Test_MustLoadFromEnvironment(t *testing.T) {
	// No need to check whether `recover()` is nil. Just turn off the panic.
	defer func() { _ = recover() }()
	var foo int
	os.Setenv("FOO", "bar")
	defer os.Clearenv()
	MustLoadFromEnvironment(&foo, "FOO")
	t.Errorf("did not panic")
}
func Test_camelToUpperSnakeCase(t *testing.T) {
	type args struct {
		s      string
		prefix string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Single uppercase",
			args: args{
				s:      "Foo",
				prefix: "",
			},
			want: "FOO",
		},
		{
			name: "Single lower",
			args: args{
				s:      "foo",
				prefix: "",
			},
			want: "FOO",
		},
		{
			name: "multi-word",
			args: args{
				s:      "FooBar",
				prefix: "",
			},
			want: "FOO_BAR",
		},
		{
			name: "prefix lower",
			args: args{
				s:      "bar",
				prefix: "foo",
			},
			want: "FOO_BAR",
		},
		{
			name: "prefix upper",
			args: args{
				s:      "bar",
				prefix: "FOO",
			},
			want: "FOO_BAR",
		},
		{
			name: "prefix only",
			args: args{
				s:      "",
				prefix: "FOO",
			},
			want: "FOO",
		},
		{
			name: "upper sequence",
			args: args{
				s:      "FooBAR",
				prefix: "",
			},
			want: "FOO_BAR",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := camelToUpperSnakeCase(tt.args.s, tt.args.prefix); got != tt.want {
				t.Errorf("camelToUpperSnakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
