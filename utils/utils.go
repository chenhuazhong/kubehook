package utils

import (
	"math/rand"
	"strings"
	"time"
)

type APIResource interface {
	String()
	Update(func(instance *APIResource, validata ...interface{}))
}

func RandString(len int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return strings.ToLower(string(bytes))
}
