package dgogm

import "fmt"

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
