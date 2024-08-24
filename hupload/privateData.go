package main

import (
	"reflect"
)

func RemovePrivateData(s any) {
	// Remove private data
	val := reflect.Indirect(reflect.ValueOf(s))

	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		typeField := val.Type().Field(i) // get field i-th of type(val)

		if f.Kind() == reflect.Struct {
			RemovePrivateData(val.Field(i).Addr().Interface())
			continue
		}

		tag := typeField.Tag.Get("scope")
		if tag == "private" {
			val.Field(i).Set(reflect.Zero(val.Field(i).Type()))
		}

		continue
	}
}
