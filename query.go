package dgogm

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/dgraph-io/dgraph/client"
	"github.com/dgraph-io/dgraph/protos"
	"github.com/media-net/cargo/logger"
)

// This struct defines a query to the dgraph
type DgQuery struct {
	s      interface{}
	id     interface{}
	fields []string
	client *client.Dgraph
}

func (dq *DgQuery) Id(id interface{}) *DgQuery {
	dq.id = id
	return dq
}

func (dq *DgQuery) Fields(fields ...string) *DgQuery {
	dq.fields = fields
	return dq
}

func (dq *DgQuery) Execute() error {
	t := reflect.TypeOf(dq.s).Elem()
	var err error
	var nodes []*protos.Node
	var qfields string
	if dq.fields == nil || len(dq.fields) == 0 {
		fields := FieldMap{}
		Debug("%s", fields.String())
		getFieldMap(t, "", fields)
		qfields = fields.String()
		goto execute
	}
	// prepare qfields from the given array
execute:
	nodes, err = query(dq.client, fmt.Sprintf(GET_NODE_FOR_ID, t.Name(), hash(GetUId(dq.s)), qfields))
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		errors.New("No results found")
	}

	parseNodeTo(nodes[0].Children[0], dq.s)
	return nil
}

// This function converts proto.Node to a map
func nodeMap(n *protos.Node) map[string]interface{} {
	m := map[string]interface{}{}
	for _, p := range n.Properties {
		v, err := convert(p.Value)
		if err != nil {
			continue
		}
		m[p.Prop] = v
	}
	for _, c := range n.Children {
		// Check if the attribute is already set
		children, ok := m[c.Attribute]
		if ok {
			switch children.(type) {
			case []*protos.Node:
				m[c.Attribute] = append(children.([]*protos.Node), c)
			case *protos.Node:
				m[c.Attribute] = []*protos.Node{children.(*protos.Node), c}
			}
			continue
		}
		m[c.Attribute] = c
	}
	return m
}

// This function parses protos.Node to fill data into given interface
func parseNodeTo(n *protos.Node, p interface{}) {
	v := reflect.ValueOf(p)
	t := reflect.TypeOf(p)
	// Fetching properties from the node
	props := nodeMap(n)
	for i := 0; i < v.Elem().NumField(); i++ {
		fname := getFieldName(t.Elem().Field(i))
		if fname == "-" {
			continue
		}
		// Search that property and assign the values
		val, ok := props[fname]
		if !ok {
			Debug("Property %s is not present in the results", fname)
			continue
		}
		Debug("%v", val)
		if !v.Elem().Field(i).IsValid() {
			logger.W("Invalid field", v.Elem().Field(i).String(), t.Elem().Field(i).Name)
			continue
		}
		Debug("%s %s", reflect.ValueOf(val).Type().String(), t.Elem().Field(i).Type)
		switch v.Elem().Field(i).Kind() {
		case reflect.Slice:
			// Check if it's slice for primitive type
			if isPrimitiveType(t.Elem().Field(i).Type.Elem()) {
				temp := []interface{}{}
				err := FromJson(val.(string), &temp)
				if err != nil {
					continue
				}
				Debug("%v %d", temp, len(temp))
				// Processing slice of primitive datatype
				// This is technically gonna be json array
				slice := reflect.MakeSlice(reflect.SliceOf(t.Elem().Field(i).Type.Elem()), 0, len(temp))
				switch t.Elem().Field(i).Type.Elem().Kind() {
				case reflect.Struct:
					Debug("Its a struct")
					for j := 0; j < len(temp); j++ {
						slice = reflect.Append(slice, reflect.ValueOf(temp[j]))
					}
				case reflect.Ptr:
					for j := 0; j < len(temp); j++ {
						ptr := reflect.New(t.Elem().Field(i).Type.Elem().Elem())
						Debug("%v %v", t.Elem().Field(i).Type.Elem().Elem().Name(), ptr)
						ptr.Elem().Set(reflect.ValueOf(temp[j]))
						Debug("Its a ptr %s")
						slice = reflect.Append(slice, ptr)
					}
				}
				Debug("%v", slice)
				v.Elem().Field(i).Set(slice)
				continue
			}
			var nodes []*protos.Node
			switch val.(type) {
			case *protos.Node:
				nodes = []*protos.Node{val.(*protos.Node)}
			case []*protos.Node:
				nodes = val.([]*protos.Node)
			}
			// Checking if the given field is already initialized
			if v.Elem().Field(i).IsNil() {
				// Initializing slice
				v.Elem().Field(i).Set(reflect.MakeSlice(reflect.SliceOf(t.Elem().Field(i).Type.Elem()), 0, len(nodes)))
			}
			Debug("Processing slices %d", len(nodes))
			// Iterating and initializing
			for j := 0; j < len(nodes); j++ {
				// Check if the elements are of type ptr, things have to be handled a bit differently
				switch t.Elem().Field(i).Type.Elem().Kind() {
				case reflect.Ptr:
					Debug("its ptr****** %d", j)
					nf := reflect.New(t.Elem().Field(i).Type.Elem().Elem())
					parseNodeTo(nodes[j], nf.Interface())
					v.Elem().Field(i).Set(reflect.Append(v.Elem().Field(i), nf))
				case reflect.Struct:
					Debug("Struct")
					Debug("%d Its a struct**** %s", j, t.Elem().Field(i).Type.Elem().String())
					nf := reflect.New(t.Elem().Field(i).Type.Elem())
					parseNodeTo(nodes[j], nf.Interface())
					v.Elem().Field(i).Set(reflect.Append(v.Elem().Field(i), nf.Elem()))
				default:
					Debug("None")
				}
			}
		case reflect.Ptr:
			if isPrimitiveType(t.Elem().Field(i).Type) {
				// This case is pointer to primitive type
				ptr := reflect.New(t.Elem().Field(i).Type.Elem())
				Debug("%v %v", t.Elem().Field(i).Type.Elem().Name(), ptr)
				ptr.Elem().Set(reflect.ValueOf(val))
				v.Elem().Field(i).Set(ptr)
				continue
			}
			nf := reflect.New(t.Elem().Field(i).Type.Elem())
			parseNodeTo(val.(*protos.Node), nf.Interface())
			v.Elem().Field(i).Set(nf)
		case reflect.Struct:
			nf := reflect.New(t.Elem().Field(i).Type)
			parseNodeTo(val.(*protos.Node), nf.Interface())
			v.Elem().Field(i).Set(nf.Elem())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float64, reflect.Float32, reflect.String, reflect.Bool, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// This includes all primitive types
			Debug("Setting %s with %v", fname, reflect.ValueOf(val))
			if v.Elem().Field(i).Type() != reflect.ValueOf(val).Type() {
				continue
			}
			v.Elem().Field(i).Set(reflect.ValueOf(val))
		}
	}
}
