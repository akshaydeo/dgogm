OGM is synonymous for ORM, Object Graph Mapping. Thanks @gazaidi for suggesting this.

## How it works?

```go
import 	"github.com/akshaydeo/dgogm"

// Struct definition
type Dog struct {
	Id        int      `dgraph:"uid"`
	Name      string   `dgraph:"name"`
	Color     *string  `json:"color" dgraph:"color"`
	Likes     []Place  `dgraph:"likes_places"`
	Nicknames []string `dgraph:"nicknames"`
	LivesAt   Place    `dgraph:"lives_at"`
	BornAt    *Place   `dgraph:"born_at"`
}

// Adding object to the dgraph
func main() {
  	dg, err := dgogm.Connect([]string{"127.0.0.1:9080"})
	if err != nil {
		t.Fail()
	}
	d := new(Dog)
	d.Id = 1
	d.Name = "jarvis"
	d.Color = dgorm.StrPtr("white")
	d.Likes = []Place{Place{1, "Pune"}, Place{2, "Mumbai"}}
	d.Nicknames = []string{"chotu", "motu"}
	d.LivesAt = Place{1, "Pune"}
	d.BornAt = &Place{3, "Solapur"}
	err = dg.Add(d)
	if err != nil {
		log.Println(err.Error())
		t.Fail()
	}
}
```

**Query:**

```graphql
{
  dog(func: eq(_xid_,"1_dog")){
    name
    color
    likes_places{
      name
    }
    nicknames
    lives_at{
      name
    }
    born_at{
      name
    }
    _xid_
  }
}
```
**Graph**

![Resulting Graph](https://github.com/akshaydeo/dgorm/raw/master/.github/one.png)

**Response**
```graphql
{
  "data": {
    "dog": [
      {
        "_uid_": "0x51a7841f167dabad",
        "name": "jarvis",
        "color": "white",
        "likes_places": [
          {
            "_uid_": "0x83a2ae6cfa98d908",
            "name": "Pune"
          },
          {
            "_uid_": "0x95197ab5b88df9a1",
            "name": "Mumbai"
          }
        ],
        "nicknames": "[\"chotu\",\"motu\"]",
        "lives_at": [
          {
            "_uid_": "0x83a2ae6cfa98d908",
            "name": "Pune"
          }
        ],
        "born_at": [
          {
            "_uid_": "0xcac342aaf75a6256",
            "name": "Solapur"
          }
        ],
        "_xid_": "1_dog"
      }
    ],
    "server_latency": {
      "total": "504µs",
      "parsing": "124µs",
      "processing": "295µs",
      "json": "82µs"
    }
  }
}
```


## Supported datatypes
- Primitive datatypes
- Pointer to struct
- Pointer to primitive datatypes
- Structs
- Slice of pointer to primitive datatypes
- Slice of primitive datatypes
- Slice of pointer to structs
- Slice of structs
