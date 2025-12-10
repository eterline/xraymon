// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.

package usecase

type float_t interface {
	~float32 | ~float64
}

type int_t interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type uint_t interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

/*
Numerable – type constraint matching all numeric types in Go.

	Used for generic IO structures to support arithmetic operations.
*/
type number interface {
	float_t | int_t | uint_t
}
