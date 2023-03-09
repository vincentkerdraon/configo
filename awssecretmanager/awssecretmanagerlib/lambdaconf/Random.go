package lambdaconf

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const lettersAlphaNum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func RandStringBytesRmndr(source string, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = source[rand.Int63()%int64(len(source))]
	}
	return string(b)
}
