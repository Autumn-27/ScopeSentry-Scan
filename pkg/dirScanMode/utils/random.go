// Package utils -----------------------------
// @file      : random.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/29 11:27
// -------------------------------------------
package utils

import (
	"math/rand"
	"time"
)

func RandomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())

	result := make([]byte, length)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}
