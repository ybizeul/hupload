package main

import "math/rand/v2"

func generateCode(groupSize int, groupsCount int) string {
	var code string

	c := "zrtypqsdfghjkmwxcvnb"
	v := "aeiouy"

	totalLength := groupSize * groupsCount
	for i := 0; i < totalLength; i++ {
		if i%groupSize == 0 && i != 0 {
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
