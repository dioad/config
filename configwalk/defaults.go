package configwalk

import (
	"reflect"
	"time"
)

// Defaults returns the defaults a config type's constructor sets, keyed by the
// keys Fields reports: the value of each scalar field and list of scalars that
// prototype, a struct or a pointer to one, holds non-zero, written as it is in
// a config file (a duration as its string, "5s"). Structs, maps, pointers and
// interfaces record no default.
func Defaults(prototype any) map[string]any {
	v := reflect.ValueOf(prototype)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	defaults := map[string]any{}
	collectDefaults(defaults, v)

	return defaults
}

func collectDefaults(into map[string]any, v reflect.Value) {
	for i := range v.NumField() {
		f, fv := v.Type().Field(i), v.Field(i)
		if IsIgnored(f) || fv.IsZero() {
			continue
		}
		if IsSquashed(f) || (f.Anonymous && f.Tag.Get("mapstructure") == "") {
			if fv.Kind() == reflect.Struct {
				collectDefaults(into, fv)
			}

			continue
		}
		if value, ok := fileValue(fv); ok {
			into[Key(f)] = value
		}
	}
}

// fileValue returns fv as it is written in a config file, if it is a scalar or
// a list of scalars.
func fileValue(fv reflect.Value) (any, bool) {
	if d, ok := reflect.TypeAssert[time.Duration](fv); ok {
		return d.String(), true
	}
	// Structs, maps, pointers and interfaces have no default recorded.
	//exhaustive:ignore
	switch fv.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return fv.Interface(), true
	case reflect.Slice:
		values := make([]any, 0, fv.Len())
		for i := range fv.Len() {
			value, ok := fileValue(fv.Index(i))
			if !ok {
				return nil, false
			}
			values = append(values, value)
		}

		return values, true
	default:
		return nil, false
	}
}
