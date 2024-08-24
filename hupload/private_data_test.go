package main

import (
	"reflect"
	"testing"
)

type TestStruct struct {
	Field1 int         `json:"field1,omitempty" scope:"private"`
	Field2 string      `json:"field2"`
	Field3 *TestStruct `json:"field3,omitempty"`
}

func TestPrivateData(t *testing.T) {
	struct1 := TestStruct{
		Field1: 1,
		Field2: "field2",
		Field3: &TestStruct{
			Field1: 1,
			Field2: "field2",
		},
	}

	RemovePrivateData(&struct1)

	want := TestStruct{
		Field2: "field2",
		Field3: &TestStruct{
			Field2: "field2",
		},
	}

	if !reflect.DeepEqual(struct1, want) {
		t.Errorf("Expected %+v, got %+v", want, struct1)
	}
}
