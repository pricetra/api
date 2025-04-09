package utils

import "math/rand"

func RangedRandomInt(min int, max int) int {
	return rand.Intn(max - min) + min
}
