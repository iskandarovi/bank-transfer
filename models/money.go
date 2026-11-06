package models

import "math"

// AmountToCents pul qiymatini aniq tiyin/sent ko'rinishidagi butun songa o'tkazadi.
// Database NUMERIC(15,2) ustunining eng katta qiymatidan oshishiga yo'l qo'ymaydi.
func AmountToCents(amount float64) (int64, bool) {
	const maxAmount = 9_999_999_999_999.99
	if amount <= 0 || amount > maxAmount || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, false
	}

	cents := math.Round(amount * 100)
	if math.Abs(amount*100-cents) > 1e-7 {
		return 0, false
	}
	return int64(cents), true
}
