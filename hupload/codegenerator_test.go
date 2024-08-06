package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestCodeGenerator(t *testing.T) {
	groupSize := 4
	groups := 3
	code := generateCode(4, 3)

	expectedLength := groupSize*groups + (groups - 1)

	if len(code) != expectedLength {
		t.Errorf("Expected code length to be %d, got %d", expectedLength, len(code))
	}

	var re_slice []string
	for i := 0; i < groups; i++ {
		re_slice = append(re_slice, buildRe(groupSize))
	}
	re := strings.Join(re_slice, "-")

	m := regexp.MustCompile(re).MatchString(code)

	if m == false {
		t.Errorf("Expected code to match %s, got %s", re, code)
	}

}

func buildRe(size int) string {
	c := "zrtypqsdfghjkmwxcvnb"
	v := "aeiouy"

	var a []string

	for i := 0; i < size; i++ {
		var s string
		if i%2 == 0 {
			s = c
		} else {
			s = v
		}
		a = append(a, fmt.Sprintf("[%s]", s))
	}
	return strings.Join(a, "")
}
