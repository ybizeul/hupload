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
			RemovePrivateData(f.Addr().Interface())
			continue
		}
		if f.Kind() == reflect.Pointer {
			if !f.IsNil() {
				RemovePrivateData(f.Elem().Addr().Interface())
			}
			continue
		}
		tag := typeField.Tag.Get("scope")
		if tag == "private" {
			if f.CanSet() {
				f.Set(reflect.Zero(val.Field(i).Type()))
			}
		}

		continue
	}
}
