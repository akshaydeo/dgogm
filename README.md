# Dgorm : Dgraph ORM


## How it works?

```go
import 	"github.com/akshaydeo/dgorm"

// Struct definition
type Dog struct {
	Id    int    `dgraph:"uid"`
	Name  string `dgraph:"name"`
	Color string `json:"color" dgraph:"color"`
}

// Adding object to the dgraph
func main() {
  	dg, err := dgorm.Connect([]string{"127.0.0.1:9080"})
  	if err != nil {
		t.Fail()
	}
	d := new(Dog)
	d.Id = 1
	d.Name = "jarvis"
	d.Color = "white"
	err = dg.Add(d)
}
```

Is mapped to
**Query:**
```graphql
{
  dog(func: eq(_xid_,"1_dog")){
    name
    color
    _xid_
  }
}
```
**Response**
```graphql
{
  "data": {
    "dog": [
      {
        "_uid_": "0x51a7841f167dabad",
        "name": "jarvis",
        "color": "white",
        "_xid_": "1_dog"
      }
    ],
    "server_latency": {
      "parsing": "135µs",
      "processing": "425µs",
      "json": "124µs",
      "total": "691µs"
    }
  }
}
```


## Supported datatypes
- Primitive datatypes
- Pointer to struct
- Structs
- Slice of pointer to primitive datatypes
- Slice of primitive datatypes
- Slice of pointer to structs
- Slice of structs
