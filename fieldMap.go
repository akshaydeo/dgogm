package dgogm

import (
	"fmt"
	"reflect"
)

// Field map type defines query fields
type FieldMap map[string][]interface{}

// This function adds an element into the fieldmap
// Along with that it maintains/updates the index field used while parsing the results
func (fm FieldMap) Add(key string, val interface{}) {
	if fm[key] != nil {
		fm[key] = append(fm[key], val)
		return
	}
	fm[key] = []interface{}{val}
	return
}

// This function converts fieldmap to query string
func (fm FieldMap) String() string {
	str := "_xid_ _uid_"
	getQuery(&str, fm)
	return str
}

// This function recursively creates query params for given field-map
func getQuery(q *string, fm FieldMap) {
	for k, v := range fm {
		if k != "" {
			*q = fmt.Sprintf("%s %s { _xid_ _uid_", *q, k)
		}
		for _, e := range v {
			switch e.(type) {
			case FieldMap:
				Debug("%s : Field map", k)
				getQuery(q, e.(FieldMap))
			case string:
				if e.(string) == "-" {
					continue
				}
				Debug("%s: %s", k, e.(string))
				*q = fmt.Sprintf("%s %s", *q, e.(string))
			}
		}
		if k != "" {
			*q = fmt.Sprintf("%s }", *q)
		}
	}
}

// This function converts types into fields query for Dgraph
func getFieldMap(t reflect.Type, parent string, m FieldMap) {
	Debug("%s", t.Name())
	for i := 0; i < t.NumField(); i++ {
		Debug("Checking if its a primitive type %s", getFieldName(t.Field(i)))
		if isPrimitiveType(t.Field(i).Type) {
			m.Add(parent, getFieldName(t.Field(i)))
			continue
		}
		Debug("Non primitive type %s", getFieldName(t.Field(i)))
		switch t.Field(i).Type.Kind() {
		case reflect.Slice:
			Debug("It's a slice")
			Debug("%s []%s", getFieldName(t.Field(i)), t.Field(i).Type.Elem().Name())
			nm := FieldMap{}
			getFieldMap(t.Field(i).Type.Elem(), getFieldName(t.Field(i)), nm)
			m.Add(parent, nm)
		case reflect.Struct:
			nm := FieldMap{}
			getFieldMap(t.Field(i).Type, getFieldName(t.Field(i)), nm)
			m.Add(parent, nm)
		case reflect.Ptr:
			Debug("It's a ptr type")
			nm := FieldMap{}
			getFieldMap(t.Field(i).Type.Elem(), getFieldName(t.Field(i)), nm)
			m.Add(parent, nm)
		}
	}
}
