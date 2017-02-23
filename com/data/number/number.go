package number

import (
	"math"
)

// 四舍五入
func Round(val float64, places int) float64 {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= 0.5 {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	return round / pow
}

// 幂函数
func Pow(x int, y int) int {
	num := 1
	for i := 0; i < y; i++ {
		num *= x
	}
	return num
}
