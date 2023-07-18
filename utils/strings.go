package utils

import (
	cryptoRand "crypto/rand"
	"math/rand"
	"strings"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
	var sb strings.Builder
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(letterBytes[rand.Intn(len(letterBytes))])
	}
	return sb.String()
}

func RandToken(n int) (string, error) {
	codes := make([]byte, n)
	if _, err := cryptoRand.Read(codes); err != nil {
		return "", err
	}

	for i := 0; i < n; i++ {
		codes[i] = uint8(48 + (codes[i] % 10))
	}

	return string(codes), nil
}

func EnsureSuffix(s string, suffix string) string {
	return strings.TrimSuffix(s, suffix) + suffix
}

func EnsurePrefix(s string, prefix string) string {
	return prefix + strings.TrimPrefix(s, prefix)
}
