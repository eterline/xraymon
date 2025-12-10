// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.
package usecase_test

import (
	"testing"

	"github.com/eterline/xraymon/internal/utils/usecase"
)

func Test_AvgVector(t *testing.T) {
	t.Run("int slice", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []int
			expected int
		}{
			{"empty slice", []int{}, 0},
			{"single element", []int{5}, 5},
			{"multiple elements", []int{1, 2, 3, 4}, 2}, // integer division
			{"negative numbers", []int{-1, -2, -3}, -2},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := usecase.AvgVector(tt.input)
				if got != tt.expected {
					t.Errorf("AvgVector(%v) = %v, want %v", tt.input, got, tt.expected)
				}
			})
		}
	})

	t.Run("float64 slice", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []float64
			expected float64
		}{
			{"empty slice", []float64{}, 0},
			{"single element", []float64{5.5}, 5.5},
			{"multiple elements", []float64{1.5, 2.5, 3.0}, 2.3333333333333335},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := usecase.AvgVector(tt.input)
				if got != tt.expected {
					t.Errorf("AvgVector(%v) = %v, want %v", tt.input, got, tt.expected)
				}
			})
		}
	})
}

func Test_AvgVectorFunc(t *testing.T) {
	type item struct {
		val int
	}

	t.Run("basic int", func(t *testing.T) {
		s := []item{{1}, {2}, {3}, {4}}
		got := usecase.AvgVectorFunc(s, func(idx int) int {
			return s[idx].val
		})
		expected := 2 // integer division
		if got != expected {
			t.Errorf("AvgVectorFunc = %v, want %v", got, expected)
		}
	})

	t.Run("float64", func(t *testing.T) {
		type fitem struct {
			v float64
		}
		s := []fitem{{1.5}, {2.5}, {3.0}}
		got := usecase.AvgVectorFunc(s, func(idx int) float64 {
			return s[idx].v
		})
		expected := (1.5 + 2.5 + 3.0) / 3
		if got != expected {
			t.Errorf("AvgVectorFunc = %v, want %v", got, expected)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		s := []item{}
		got := usecase.AvgVectorFunc(s, func(idx int) int {
			return 1
		})
		if got != 0 {
			t.Errorf("AvgVectorFunc(empty) = %v, want 0", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		s := []item{{42}}
		got := usecase.AvgVectorFunc(s, func(idx int) int {
			return s[idx].val
		})
		if got != 42 {
			t.Errorf("AvgVectorFunc(single) = %v, want 42", got)
		}
	})
}

func Test_PercentOfOverfull(t *testing.T) {
	type TCase struct {
		full     any
		frac     any
		expected float64
	}

	tests := []TCase{
		{100, 50, 50},
		{100, 100, 100},
		{100, 150, 150},
		{200, 25, 12.5},
		{uint(100), uint(40), 40},
		{uint(100), uint(200), 200},
		{float64(8), float64(1), 12.5},
	}

	for _, tt := range tests {
		var got float64

		switch full := tt.full.(type) {
		case int:
			got = usecase.PercentOfOverfull(full, tt.frac.(int))
		case uint:
			got = usecase.PercentOfOverfull(full, tt.frac.(uint))
		case float64:
			got = usecase.PercentOfOverfull(full, tt.frac.(float64))
		default:
			t.Fatalf("unsupported type in test")
		}

		if got != tt.expected {
			t.Errorf("PercentOfOverfull(%v, %v) = %v; want %v",
				tt.full, tt.frac, got, tt.expected)
		}
	}
}

func Test_PercentOf(t *testing.T) {
	type TCase struct {
		full     any
		frac     any
		expected float64
	}

	tests := []TCase{
		{100, 50, 50},
		{100, 100, 100},
		{100, 150, 100}, // clamp
		{100, 999, 100}, // clamp
		{8, 1, 12.5},
		{uint(100), uint(200), 100},
		{float64(100), float64(150), 100},
	}

	for _, tt := range tests {
		var got float64

		switch full := tt.full.(type) {
		case int:
			got = usecase.PercentOf(full, tt.frac.(int))
		case uint:
			got = usecase.PercentOf(full, tt.frac.(uint))
		case float64:
			got = usecase.PercentOf(full, tt.frac.(float64))
		default:
			t.Fatalf("unsupported type in test")
		}

		if got != tt.expected {
			t.Errorf("PercentOf(%v, %v) = %v; want %v",
				tt.full, tt.frac, got, tt.expected)
		}
	}
}
