package tools

import (
	"fmt"
	"math/rand"
	"testing"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr(n int) string {
	if n == 0 {
		n = 1
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestGenerateRsa(t *testing.T) {
	for i := 0; i <= 1000; i++ {
		rsa, err := GenerateRsa(randStr(rand.Intn(20)))
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(rsa)
	}
}
