package dgorm

import (
	"encoding/json"
	"errors"

	"fmt"
	"reflect"
	"strings"
	"time"

	"hash/fnv"

	"github.com/dgraph-io/dgraph/client"
	"github.com/dgraph-io/dgraph/protos"
)

// This function returns the name for the given struct field
// First dgraph -> then json -> then field name
func getFieldName(f reflect.StructField) string {
	val, ok := f.Tag.Lookup("dgraph")
	if ok {
		return val
	}
	val, ok = f.Tag.Lookup("json")
	if ok {
		return val
	}
	return f.Name
}

// This function returns a new 64-bit FNV-1a hash.Hash
func hash(i string) uint64 {
	f := fnv.New64a()
	f.Write([]byte(i))
	return f.Sum64()
}

// This function returns _uid_ for the given struct
// It checks if there is a function called UId which returns string,
// if it's there, that function will be used to return the uid
func GetUId(p interface{}) string {
	// Get type info of p
	t := reflect.TypeOf(p)
	if t.Kind() != reflect.Ptr {
		panic("GetUId expects pointer to struct")
	}
	_, ok := t.MethodByName("UId")
	if ok {
		uid := reflect.ValueOf(p).MethodByName("UId").Call([]reflect.Value{})[0]
		switch uid.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Bool, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return fmt.Sprintf("%d", uid.Int())
		case reflect.Float64, reflect.Float32:
			return fmt.Sprintf("%f", uid.Float())
		case reflect.String:
			return uid.String()
		}
	}
	for i := 0; i < t.Elem().NumField(); i++ {
		if getFieldName(t.Elem().Field(i)) == "uid" {
			switch t.Elem().Field(i).Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Bool, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return fmt.Sprintf("%d_%s",
					reflect.ValueOf(p).Elem().FieldByName(t.Elem().Field(i).Name).Int(),
					strings.ToLower(t.Elem().Name()))
			case reflect.Float64, reflect.Float32:
				return fmt.Sprintf("%f_%s",
					reflect.ValueOf(p).Elem().FieldByName(t.Elem().Field(i).Name).Float(),
					strings.ToLower(t.Elem().Name()))
			case reflect.String:
				return fmt.Sprintf("%s_%s",
					reflect.ValueOf(p).Elem().FieldByName(t.Elem().Field(i).Name).String(),
					strings.ToLower(t.Elem().Name()))
			}
		}
	}
	return ""
}

// Get the corresponding protos.Value object for the given interface
func getVal(val interface{}) *protos.Value {
	switch val.(type) {
	case int:
		return &protos.Value{Val: &protos.Value_IntVal{IntVal: int64(val.(int))}}
	case string:
		if val.(string) == "" {
			return nil
		}
		return &protos.Value{Val: &protos.Value_StrVal{StrVal: strings.Replace(val.(string), "\"", "\\\"", -1)}}
	case float64:
		return &protos.Value{Val: &protos.Value_DoubleVal{DoubleVal: val.(float64)}}
	case time.Time:
		return &protos.Value{Val: &protos.Value_DatetimeVal{DatetimeVal: []byte(val.(time.Time).String())}}
	case bool:
		return &protos.Value{Val: &protos.Value_BoolVal{BoolVal: val.(bool)}}
	case []byte:
		return &protos.Value{Val: &protos.Value_BytesVal{BytesVal: val.([]byte)}}
	case *GeoPoint:
		return &protos.Value{Val: &protos.Value_GeoVal{GeoVal: []byte(*(val.(*GeoPoint).Json()))}}
	}
	return nil
}

// Sets val to the given edge
func setVal(edge *client.Edge, val interface{}) error {
	switch val.(type) {
	case int, int64, int8, int32, int16:
		return edge.SetValueInt(val.(int64))
	case string:
		if val.(string) == "" {
			return errors.New("Empty")
		}
		return edge.SetValueString(strings.Replace(val.(string), "\"", "\\\"", -1))
	case float64, float32:
		return edge.SetValueFloat(val.(float64))
	case time.Time:
		return edge.SetValueDatetime(val.(time.Time))
	case bool:
		return edge.SetValueBool(val.(bool))
	case []byte:
		return edge.SetValueBytes(val.([]byte))
	case *GeoPoint:
		return edge.SetValueGeoJson(*(val.(*GeoPoint).Json()))
	}
	return errors.New("Val type is not supported ")
}

// This function gives pointer to json string of the struct, ignoring the errors
func ToJsonUnsafe(v interface{}) *string {
	data, err := json.Marshal(v)
	if err != nil {
		Error("Error while converting to json", err)
	}
	return StrPtr(string(data))
}

// Function to return pointer to string in the param, just easier way to do this
func StrPtr(s string) *string {
	return &s
}
