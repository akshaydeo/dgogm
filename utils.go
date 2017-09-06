package dgogm

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
	"github.com/media-net/cargo/logger"
	"github.com/satori/go.uuid"
)

// This function Checks if the given value is zero or not
// https://github.com/golang/go/issues/7501#issuecomment-66092219
func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}
	return false
}

// This function converts protos.Value to corresponding goland datatype
func convert(val *protos.Value) (interface{}, error) {
	switch val.Val.(type) {
	case *protos.Value_StrVal:
		return strings.Replace(val.GetStrVal(), "\\\"", "\"", -1), nil
	case *protos.Value_BoolVal:
		return val.GetBoolVal(), nil
	case *protos.Value_IntVal:
		return val.GetIntVal(), nil
	case *protos.Value_UidVal:
		return val.GetUidVal(), nil
	case *protos.Value_DateVal:
		logger.D("Parsing time")
		t, err := time.Parse("2006-01-02 19:54:00.000000000 +0000 UTC", string(val.GetDatetimeVal()))
		if err != nil {
			logger.E("Error while parsing datetime", err)
			return string(val.GetDatetimeVal()), err
		}
		return t, nil
	case *protos.Value_DatetimeVal:
		logger.D("Parsing time")
		t, err := time.Parse("2006-01-02 19:54:00.000000000 +0000 UTC", string(val.GetDatetimeVal()))
		if err != nil {
			logger.E("Error while parsing datetime", err)
			return string(val.GetDatetimeVal()), err
		}
		return t, nil
	default:
		return val.GetDefaultVal(), nil
	}
	return nil, nil
}

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
	// If there is no provided id, then generating random uuid
	return uuid.NewV4().String()
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
func ToJsonUnsafe(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		Error("Error while converting to json", err)
	}
	return string(data)
}

// This function converts json string to interface
func FromJson(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

// Function to return pointer to string in the param, just easier way to do this
func StrPtr(s string) *string {
	return &s
}
