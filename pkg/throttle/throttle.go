package throttle

import (
	"log"
	"strconv"
	"unicode"

	"golang.org/x/time/rate"
)

func NewLimiter(rateStr string) *rate.Limiter {
	limitRate, err := ParseLimitRate(rateStr)
	if err != nil {
		log.Fatal("Error: limit formatting is wrong")
	}

	return rate.NewLimiter(rate.Limit(limitRate), limitRate)
}

func ParseLimitRate(rateStr string) (int, error) {
	var multiplier int
	if len(rateStr) > 1 {
		unit := unicode.ToLower(rune(rateStr[len(rateStr)-1]))
		switch unit {
		case 'k':
			multiplier = 1024
		case 'm':
			multiplier = 1024 * 1024
		case 'g':
			multiplier = 1024 * 1024 * 1024
		}

		rateStr = rateStr[:len(rateStr)-1]
	}

	value, err := strconv.Atoi(rateStr)
	if err != nil {
		return 0, err
	}

	return value * multiplier, nil
}
