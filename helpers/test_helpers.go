package helpers

import (
	"math/rand"
	"strings"
)

func RandomName() string {
	chars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	length := 16
	var builder strings.Builder

	for i := 0; i < length; i++ {
		builder.WriteRune(chars[rand.Intn(len(chars))])
	}

	return builder.String()
}
