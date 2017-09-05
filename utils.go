package dgorm

import (
	"encoding/json"

	"fmt"
	"reflect"
	"strings"
	"time"

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

// This function returns _uid_ for the given struct
func GetUId(p interface{}) string {
	// Get type info of p
	t := reflect.TypeOf(p)
	if t.Kind() != reflect.Ptr {
		panic("GetUId expects pointer to struct")
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

func ToJsonUnsafe(v interface{}) *string {
	data, err := json.Marshal(v)
	if err != nil {
		Error("Error while converting to json", err)
	}
	return StrPtr(string(data))
}

func StrPtr(s string) *string {
	return &s
}
