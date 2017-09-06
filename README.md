# Dgogm : Dgraph OGM

OGM is synonymous for ORM, Object Graph Mapping. Thanks @gazaidi for suggesting this.

## How it works?
### Struct for this example
```go
type Dog struct {
	Id        int      `dgraph:"uid"`
	Name      string   `dgraph:"name"`
	Color     *string  `json:"color" dgraph:"color"`
	Likes     []Place  `dgraph:"likes_places"`
	Nicknames []string `dgraph:"nicknames"`
	LivesAt   Place    `dgraph:"lives_at"`
	BornAt    *Place   `dgraph:"born_at"`
}
```
### Adding struct to the graph
#### Using existing client connection
```go
// Adding object to the dgraph
func main() {
  	var client *client.Dgraph
	// perform client connection	
	d := new(Dog)
	d.Id = 1
	d.Name = "jarvis"
	d.Color = dgorm.StrPtr("white")
	d.Likes = []Place{Place{1, "Pune"}, Place{2, "Mumbai"}}
	d.Nicknames = []string{"chotu", "motu"}
	d.LivesAt = Place{1, "Pune"}
	d.BornAt = &Place{3, "Solapur"}
	err = dgogm.Add(c, d)
	if err != nil {
		log.Println(err.Error())
		t.Fail()
	}
}
```
#### Using dgogm client 
```go
// Adding object to the dgraph
func main() {
	dg, err := dgogm.Connect([]string{"127.0.0.1:9080"})
	if err != nil {
		t.Fail()
	}
	d := new(Dog)
	d.Id = 1
	d.Name = "jarvis"
	d.Color = dgogm.StrPtr("white")
	d.Likes = []Place{Place{1, "Pune"}, Place{2, "Mumbai"}}
	d.Nicknames = []string{"chotu", "motu"}
	d.LivesAt = Place{1, "Pune"}
	d.BornAt = &Place{3, "Solapur"}
	err = dg.Add(d)
}
```
**Resultant Graph Data**

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
### Getting structs back from the DGraph
#### Using dgog client
```go
func main() {
	dg, err := dgogm.Connect([]string{"127.0.0.1:9080"})
	if err != nil {
		t.Fail()
	}
	d := new(Dog)
	d.Id = 1
	err = dg.Find(d).Id(1).Execute()
	if err != nil {
		t.Fail()
	}
	fmt.Printf("%v", d)
}
```
**Output**
```bash
&{1 jarvis 0xc42021b270 [{0 Pune} {0 Mumbai}] [] {0 Pune} 0xc4200f9cc0}
```
#### Using existing client
```go
func main() {
	var c *client.Dgraph
	// make connection
	d := new(Dog)
	d.Id = 1
	err = Find(c, d).Id(1).Execute()
	if err != nil {
		t.Fail()
	}
	fmt.Printf("%v", d)
}
```
**Output**
```bash
&{1 jarvis 0xc42021b270 [{0 Pune} {0 Mumbai}] [] {0 Pune} 0xc4200f9cc0}
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
