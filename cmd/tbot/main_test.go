package main

import (
	"strconv"
	"strings"
	"testing"
)

func Test_validChats(t *testing.T) {
	goodChats := "-125646 569854 565874 -46546 0 1"
	badChats := "11111 this is bad chat string 1111111"
	splitted := strings.Split(goodChats, " ")
	validChats, err := parseValidChats(goodChats)
	if err != nil {
		t.Fatalf("%v: parsing: %v\n", goodChats, err)
	}

	lengthTestFunc := func(t *testing.T) {
		if len(splitted) != len(validChats) {
			t.Fatalf("expected length of parsed values from '%s' will be %d got %d\n",
				goodChats, len(splitted), len(validChats))
		}
	}

	t.Run("length", lengthTestFunc)

	t.Run("good_input", func(t *testing.T) {

		lengthTestFunc(t)

		for _, v := range splitted {
			n, _ := strconv.ParseInt(v, 10, 0)
			if !validChats[n] {
				t.Fatalf("expected parsed value '%d' will be in '%v'\n", n, validChats)
			}
		}
	})

	t.Run("bad_input", func(t *testing.T) {
		_, err := parseValidChats(badChats)
		if err == nil {
			t.Fatalf("expected error while parsing '%v', got nil instead\n", badChats)
		}

	})

}
