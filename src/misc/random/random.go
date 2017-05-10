package random

import (
	"math/rand"
)

func RandInt32(min int32, max int32) int32 {
	if max <= min {
		max = min + 1
	}
	var base int32 = 0
	if min < 0 {
		base = -min
		min += base
		max += base
	}
	return -base + min + rand.Int31n(max-min)
}
func RandInt64(min int64, max int64) int64 {
	if max <= min {
		max = min + 1
	}
	var base int64 = 0
	if min < 0 {
		base = -min
		min += base
		max += base
	}
	return -base + min + rand.Int63n(max-min)
}
func RandInt(min int, max int) int {
	if max <= min {
		max = min + 1
	}
	var base int = 0
	if min < 0 {
		base = -min
		min += base
		max += base
	}
	return -base + min + rand.Intn(max-min)
}
