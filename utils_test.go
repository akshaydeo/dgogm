package dgogm_test

import (
	"testing"

	"log"

	"github.com/akshaydeo/dgogm"
)

type Person1 struct {
	Id       string `dgraph:"uid"`
	Name     string `dgraph:"name"`
	DoesCode bool   `dgraph:"codes"`
	Height   int    `dgraph:"height_in_cm"`
}

type Person2 struct {
	Id       int64  `dgraph:"uid"`
	Name     string `dgraph:"name"`
	DoesCode bool   `dgraph:"codes"`
	Height   int    `dgraph:"height_in_cm"`
}

type Person3 struct {
	Id       int64  `dgraph:"uid"`
	Name     string `dgraph:"name"`
	DoesCode bool   `dgraph:"codes"`
	Height   int    `dgraph:"height_in_cm"`
}

func (p *Person3) UId() string {
	return "test"
}

func TestGetUId(t *testing.T) {
	s := Person1{
		"this_is_test",
		"Akshay Deo",
		true,
		183,
	}
	if dgogm.GetUId(&s) != "this_is_test_person1" {
		t.Fail()
	}
}

func TestGetUIdForInt64Id(t *testing.T) {
	s := Person2{
		12312314,
		"Akshay Deo",
		true,
		183,
	}
	if dgogm.GetUId(&s) != "12312314_person2" {
		t.Fail()
	}
}

func TestGetUIdForUIdFunc(t *testing.T) {
	s := Person3{
		12312314,
		"Akshay Deo",
		true,
		183,
	}
	log.Println(dgogm.GetUId(&s))
	if dgogm.GetUId(&s) != "test" {
		t.Fail()
	}
}
