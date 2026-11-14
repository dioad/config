package configwalk

import (
	"reflect"
	"slices"
	"strings"
)

// Key returns the key f is read from: its mapstructure name, or its Go name if
// it has none, lower-cased.
func Key(f reflect.StructField) string {
	name, _, _ := strings.Cut(f.Tag.Get("mapstructure"), ",")
	if name == "" {
		name = f.Name
	}

	return strings.ToLower(name)
}

// IsSquashed reports whether f's fields are read from its parent's keys.
func IsSquashed(f reflect.StructField) bool {
	_, options, _ := strings.Cut(f.Tag.Get("mapstructure"), ",")

	return slices.Contains(strings.Split(options, ","), "squash")
}

// IsIgnored reports whether f is not a key: it is unexported and not embedded,
// or tagged `mapstructure:"-"`.
func IsIgnored(f reflect.StructField) bool {
	name, _, _ := strings.Cut(f.Tag.Get("mapstructure"), ",")

	return (!f.IsExported() && !f.Anonymous) || name == "-"
}

// Elem returns the type a field of type t holds values of, looking through
// pointers, lists and maps.
func Elem(t reflect.Type) reflect.Type {
	for {
		// Only container kinds are looked through.
		//exhaustive:ignore
		switch t.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
			t = t.Elem()
		default:
			return t
		}
	}
}

// Field is one key of a struct.
type Field struct {
	// Key is the key the field is read from.
	Key string
	// StructField is the field itself. For a field of a squashed or embedded
	// struct it is that struct's field, not the field that holds the struct.
	StructField reflect.StructField
	// Owner is the struct type StructField belongs to.
	Owner reflect.Type
}

// Fields returns the keys of struct type t, in field order. The fields of an
// embedded struct, and of a struct tagged ",squash", are inlined, since they
// are read from t's own keys; a field ignored by IsIgnored is left out.
func Fields(t reflect.Type) []Field {
	t = Elem(t)
	if t.Kind() != reflect.Struct {
		return nil
	}

	var fields []Field
	for i := range t.NumField() {
		f := t.Field(i)
		if IsIgnored(f) {
			continue
		}
		if inner := Elem(f.Type); inner.Kind() == reflect.Struct && (IsSquashed(f) || (f.Anonymous && f.Tag.Get("mapstructure") == "")) {
			fields = append(fields, Fields(inner)...)

			continue
		}
		fields = append(fields, Field{Key: Key(f), StructField: f, Owner: t})
	}

	return fields
}

// SquashedFields returns t's named struct fields tagged ",squash", including
// those of the structs it embeds. Some reflectors inline only embedded structs,
// so a generator can use these to inline the named ones as well.
func SquashedFields(t reflect.Type) []reflect.StructField {
	var fields []reflect.StructField
	for i := range Elem(t).NumField() {
		f := Elem(t).Field(i)
		if !f.IsExported() && !f.Anonymous {
			continue
		}
		inner := Elem(f.Type)
		if f.Anonymous && (IsSquashed(f) || f.Tag.Get("mapstructure") == "") {
			if inner.Kind() == reflect.Struct {
				fields = append(fields, SquashedFields(inner)...)
			}

			continue
		}
		if IsSquashed(f) && inner.Kind() == reflect.Struct {
			fields = append(fields, f)
		}
	}

	return fields
}

// Walk calls visit once for each struct type reachable from root through
// fields, pointers, lists and maps, with the struct's keys as Fields returns
// them. A struct type that refers to itself is visited once.
func Walk(root reflect.Type, visit func(t reflect.Type, fields []Field)) {
	seen := map[reflect.Type]bool{}
	var walk func(reflect.Type)
	walk = func(t reflect.Type) {
		t = Elem(t)
		if t.Kind() != reflect.Struct || seen[t] {
			return
		}
		seen[t] = true

		visit(t, Fields(t))
		for i := range t.NumField() {
			f := t.Field(i)
			if !IsIgnored(f) {
				walk(f.Type)
			}
		}
	}
	walk(root)
}
