package models

import "testing"

func TestAmountToCents(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  int64
		valid bool
	}{
		{"whole amount", 10, 1000, true},
		{"two decimal amount", 0.29, 29, true},
		{"zero", 0, 0, false},
		{"negative", -1, 0, false},
		{"three decimals", 1.234, 0, false},
		{"too large", 10_000_000_000_000, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, valid := AmountToCents(tt.value)
			if valid != tt.valid || got != tt.want {
				t.Fatalf("AmountToCents(%v) = (%d, %v), want (%d, %v)", tt.value, got, valid, tt.want, tt.valid)
			}
		})
	}
}
