package value

import (
	"reflect"
)

// IsZero checks if the value is nil or empty or zero
func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		zero := true
		for i := 0; i < v.Len(); i++ {
			zero = zero && IsZero(v.Index(i))
		}
		return zero
	case reflect.Struct:
		zero := true
		for i := 0; i < v.NumField(); i++ {
			zero = zero && IsZero(v.Field(i))
		}
		return zero
	}
	// Compare other types directly:
	zero := reflect.Zero(v.Type())
	return v.Interface() == zero.Interface()
}
