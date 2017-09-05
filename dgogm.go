package dgogm

import (
	"reflect"

	"context"

	"fmt"

	"github.com/dgraph-io/dgraph/client"
	"github.com/dgraph-io/dgraph/protos"
	"github.com/media-net/cargo/logger"
	"github.com/pkg/errors"
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
	_, err := d.add(GetUId(p), p)
	return err
}

// This function finds the given node from the Dgraph using given id
func (d *Dgraph) FindById(p interface{}) error {
	t := reflect.TypeOf(p).Elem()
	fields := FieldMap{}
	getFieldMap(t, "", fields)
	Debug("%s", fields.String())
	nodes, err := d.query(fmt.Sprintf(GET_NODE_FOR_ID, t.Name(), hash(GetUId(p)), fields.String()))
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		errors.New("No results found")
	}
	parseNodeTo(nodes[0].Children[0], p)
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
			var nodes []*protos.Node
			switch val.(type) {
			case *protos.Node:
				nodes = []*protos.Node{val.(*protos.Node)}
			case []*protos.Node:
				nodes = val.([]*protos.Node)
			}
			// Check if types match or not
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

// This function fires the given query on connected dgraph and fetches back the response
// Does no alteration to the response
func (d *Dgraph) query(q string) ([]*protos.Node, error) {
	req := new(client.Req)
	Debug("Firing %s", q)
	req.SetQuery(q)
	resp, err := d.client.Run(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return resp.N, err
}

// Internal function, performing addition of the object into dgraph
func (d *Dgraph) add(sid string, p interface{}) (*client.Node, error) {
	// Get type info of p
	t := reflect.TypeOf(p)
	// Get value info of p
	v := reflect.ValueOf(p)
	Debug("sid is %s", sid)
	Debug("------\n %v", p, t.String(), v.String())
	var err error
	// Creating request object
	r := new(client.Req)
	// Creating source node and process _xid_ to it
	snode := d.client.NodeUid(hash(sid))
	e := snode.Edge("_xid_")
	e.SetValueString(sid)
	err = r.Set(e)
	if err != nil {
		return nil, err
	}
	// Ranging over the interface fields
	for i := 0; i < v.Elem().NumField(); i++ {
		fname := getFieldName(t.Elem().Field(i))
		if fname == "-" {
			continue
		}
		// Skip zero values
		if IsZero(v.Elem().Field(i)) {
			continue
		}
		Debug("Adding edge %s", getFieldName(t.Elem().Field(i)))
		switch v.Elem().Field(i).Kind() {
		case reflect.Slice:
			var tnode *client.Node
			if v.Elem().Field(i).Len() == 0 {
				return nil, nil
			}
			// Check if this array contains a primitive kind of elements
			if isPrimitiveType(v.Elem().Field(i).Index(0).Type()) {
				// Then jsonify them and push them inside
				Debug("Adding %s", ToJsonUnsafe(v.Elem().Field(i).Interface()))
				_, err = d.process(r, snode, t.Elem().Field(i), reflect.ValueOf(ToJsonUnsafe(v.Elem().Field(i).Interface())))
				if err != nil {
					return nil, err
				}
				continue
			}
			for j := 0; j < v.Elem().Field(i).Len(); j++ {
				switch v.Elem().Field(i).Index(j).Kind() {
				case reflect.Struct:
					tnode, err = d.add(GetUId(v.Elem().Field(i).Index(j).Addr().Interface()),
						v.Elem().Field(i).Index(j).Addr().Interface())
					if err != nil {
						return nil, err
					}
					e = snode.ConnectTo(getFieldName(t.Elem().Field(i)), *tnode)
					err = r.Set(e)
					if err != nil {
						return nil, err
					}
				case reflect.Ptr:
					tnode, err = d.add(GetUId(v.Elem().Field(i).Index(j).Interface()),
						v.Elem().Field(i).Index(j).Interface())
					if err != nil {
						return nil, err
					}
					if tnode == nil {
						continue
					}
					e = snode.ConnectTo(getFieldName(t.Elem().Field(i)), *tnode)
					err = r.Set(e)
					if err != nil {
						return nil, err
					}
				default:
					return nil, errors.New("Does not support " + t.Elem().Field(i).Type.String())
				}
			}
		default:
			_, err = d.process(r, snode, t.Elem().Field(i), v.Elem().Field(i))
			if err != nil {
				return nil, err
			}
		}
	}
	_, err = d.client.Run(context.Background(), r)
	if err != nil {
		return nil, err
	}
	return &snode, nil
}

// This function does the core processing of the fields
// Detects the name of the field, type of the field, and decides how to attach it with
// all the available information
func (d *Dgraph) process(r *client.Req, snode client.Node, field reflect.StructField, value reflect.Value) (*client.Edge, error) {
	var e client.Edge
	var err error
	switch value.Kind() {
	case reflect.Ptr:
		Debug(value.Elem().Kind().String())
		// Checking if its pointer to primitve data type
		if isPrimitiveType(value.Elem().Type()) {
			// its pointer to primitive kind
			return d.process(r, snode, field, value.Elem())
		}
		Debug("its ptr******", field.Type.Elem())
		tnode, err := d.add(GetUId(value.Interface()), value.Interface())
		if err != nil {
			return nil, err
		}
		if tnode == nil {
			return nil, nil
		}
		e = snode.ConnectTo(getFieldName(field), *tnode)
		err = r.Set(e)
		if err != nil {
			return nil, err
		}
	case reflect.Struct:
		Debug("its struct******", field.Type)
		tnode, err := d.add(GetUId(value.Addr().Interface()), value.Addr().Interface())
		if err != nil {
			return nil, err
		}
		if tnode == nil {
			return nil, nil
		}
		e = snode.ConnectTo(getFieldName(field), *tnode)
		err = r.Set(e)
		if err != nil {
			return nil, err
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e = snode.Edge(getFieldName(field))
		err = setVal(&e, value.Int())
		if err != nil {
			return nil, err
		}
		err = r.Set(e)
		if err != nil {
			return nil, err
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// This includes all primitive types
		e = snode.Edge(getFieldName(field))
		err = setVal(&e, value.Uint())
		if err != nil {
			return nil, err
		}
		err = r.Set(e)
		if err != nil {
			return nil, err
		}
	case reflect.Float64, reflect.Float32:
		e = snode.Edge(getFieldName(field))
		err = setVal(&e, value.Float())
		if err != nil {
			return nil, err
		}
		err = r.Set(e)
		if err != nil {
			return nil, err
		}
	case reflect.String:
		Debug("Adding string %s => %s", getFieldName(field), value.String())
		e = snode.Edge(getFieldName(field))
		err = setVal(&e, value.String())
		if err != nil {
			return nil, err
		}
		err = r.Set(e)
		if err != nil {
			return nil, err
		}
	case reflect.Bool:
		e = snode.Edge(getFieldName(field))
		err = setVal(&e, value.Bool())
		if err != nil {
			return nil, err
		}
		err = r.Set(e)
		if err != nil {
			return nil, err
		}
	}
	return &e, nil
}
