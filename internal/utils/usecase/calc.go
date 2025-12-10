// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.

package usecase

func AvgVector[T float_t | int_t | uint_t](s []T) T {
	ln := len(s)

	if ln == 0 {
		return 0
	}

	if ln == 1 {
		return s[0]
	}

	sum := T(0)

	for _, v := range s {
		sum += v
	}

	return sum / T(ln)
}

func AvgVectorFunc[V any, T float_t | int_t | uint_t](s []V, idxItemFunc func(idx int) T) T {
	ln := len(s)

	if ln == 0 {
		return 0
	}

	if ln == 1 {
		return idxItemFunc(0)
	}

	sum := T(0)

	for i := 0; i < ln; i++ {
		sum += idxItemFunc(i)
	}

	return sum / T(ln)
}

func PercentOfOverfull[T float_t | int_t | uint_t](full, frac T) float64 {
	if full == 0 {
		return 0
	}
	return float64(frac) / float64(full) * 100.0
}

func PercentOf[T float_t | int_t | uint_t](full, frac T) float64 {
	pct := PercentOfOverfull(full, frac)
	if pct > 100.0 {
		pct -= pct - 100.0
	}
	return pct
}
