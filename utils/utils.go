package utils

import (
  "strings"
	"math/rand"
)

// RandInt generate random int in range [min, max].
func RandInt(min int, max int) int {
	return rand.Intn(max-min) + min
}

func DetermineCharset(contentType string) string {
  if contentType != "" {
    s := strings.Split(contentType, ";")
    charset := strings.TrimSpace(s[len(s)-1])
    if strings.HasPrefix(charset, "charset=") {
      return charset[8:len(charset)]
    }
  }
  return "utf-8"
}
