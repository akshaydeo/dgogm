# Dgorm : Dgraph ORM


## How it works?

```go
type Dog struct {
	Id    int    `dgraph:"uid"`
	Name  string `dgraph:"name"`
	Color string `json:"color" dgraph:"color"`
}
```

Is mapped to
```
```
