package main

import "math/rand/v2"

func generateCode() string {
	code := ""

	c := "zrtypqsdfghjkmwxcvnb"
	v := "aeiouy"

	for i := 0; i < 12; i++ {
		if i%4 == 0 && i != 0 {
			code += "-"
		}
		if i%2 == 0 {
			code += string(c[rand.IntN(len(c))])
		} else {
			code += string(v[rand.IntN(len(v))])
		}
	}
	return code
}
