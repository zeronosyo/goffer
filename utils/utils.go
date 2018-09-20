package utils

import (
	"math/rand"
)

// RandInt generate random int in range [min, max].
func RandInt(min int, max int) int {
	return rand.Intn(max-min) + min
}
