package dgorm_test

import (
	"testing"

	"github.com/akshaydeo/dgorm"
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

func TestGetUIdForStringId(t *testing.T) {
	s := Person1{
		"this_is_test",
		"Akshay Deo",
		true,
		183,
	}
	if dgorm.GetUId(&s) != "this_is_test_person1" {
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
	if dgorm.GetUId(&s) != "12312314_person2" {
		t.Fail()
	}
}
