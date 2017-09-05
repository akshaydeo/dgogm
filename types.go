package dgogm

import "reflect"

// This function returns if given type is a or points to a primitive type
func isPrimitiveType(tp reflect.Type) bool {
	switch tp.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float64, reflect.Float32, reflect.String, reflect.Bool:
		return true
	case reflect.Ptr:
		return isPrimitiveType(tp.Elem())
	}
	return false
}
