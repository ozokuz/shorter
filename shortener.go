package main

import "math/rand"

const CHARS string = "abcdefghijklmnopqrstuvwxyz0123456789"

func Shorten() string {
	str := ""

	for i := 0; i < 8; i++ {
		str += string(CHARS[rand.Intn(len(CHARS))])
	}

	return str
}
