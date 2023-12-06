package luna

import (
	"strconv"
)

func Valid(number string) bool {
	var sum int
	parity := len(number) % 2
	for i, digit := range number {
		if d, err := strconv.Atoi(string(digit)); err == nil {
			if i%2 == parity {
				d *= 2
				if d > 9 {
					d -= 9
				}
			}
			sum += d
		} else {
			return false
		}
	}
	return sum%10 == 0
}
