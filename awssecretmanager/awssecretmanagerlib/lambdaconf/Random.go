package lambdaconf

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const lettersAlphaNum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// RandStringBytesRmndr generates a password
func RandStringBytesRmndr(source string, n int) string {
	//Not using the best implementation from
	//https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
	//but seems good enough

	b := make([]byte, n)
	for i := range b {
		b[i] = source[rand.Int63()%int64(len(source))]
	}
	return string(b)
}
