package dgorm_test

import (
	"testing"

	"log"

	"github.com/akshaydeo/dgorm"
)

type Dog struct {
	Id    int    `dgraph:"uid"`
	Name  string `dgraph:"name"`
	Color string `json:"color" dgraph:"color"`
}

func TestDgraph_Add(t *testing.T) {
	dg, err := dgorm.Connect([]string{"127.0.0.1:9080"})
	if err != nil {
		t.Fail()
	}
	d := new(Dog)
	d.Id = 1
	d.Name = "jarvis"
	d.Color = "white"
	err = dg.Add(d)
	if err != nil {
		log.Println(err.Error())
		t.Fail()
	}
}
