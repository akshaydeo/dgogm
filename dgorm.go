package dgorm

import (
	"reflect"
)

// This function adds the given pointer to struct into the Dgraph
// The rules are as follows:
// 1. This library handles the _uid_ and _xid_ creation
// 2. UID is created by using `struct name + id field` in the struct and creating a hash out of it
// 3. XID is created by fmt.Sprintf("%s_%s",id,struct_name)
// 4. This library looks for dgraph tags for the field, if they are not available, they go for JSON tags, if that is not available it goes for field names
// 5. If the field is a primitive type, its added as a predicate to the given node
// 6. If the field is a struct or pointer to struct then a new relation node is added
func (d *Dgraph) Add(p interface{}) error {
	// Get type info of p
	t := reflect.TypeOf(p)
	// Get value info of p
	v := reflect.ValueOf(p)
	Debug("------\n", p, t.String(), v.String())
	// Ranging over the interface fields
	for i := 0; i < v.Elem().NumField(); i++ {
		fname := getFieldName(t.Elem().Field(i))
		if fname == "-" {
			continue
		}
		Debug("Adding", v.Elem().Field(i).String(), t.Elem().Field(i).Type)
		switch v.Elem().Field(i).Kind() {
		case reflect.Slice:
			for j := 0; j < v.Elem().Field(i).Len(); j++ {
				Debug("loop", j, v.Elem().Field(i).Type(), t.Elem().Field(i).Type.Elem())
				// Check if the elements are of type ptr, things have to be handled a bit differently
				switch t.Elem().Field(i).Type.Elem().Kind() {
				case reflect.Ptr:
					Debug("its ptr******")

				case reflect.Struct:

				}
			}
		case reflect.Ptr:
			Debug("its ptr******", t.Elem().Field(i).Type.Elem())

		case reflect.Struct:
			Debug("its struct******", t.Elem().Field(i).Type)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float64, reflect.Float32, reflect.String, reflect.Bool, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// This includes all primitive types

		}
	}
	return nil
}
