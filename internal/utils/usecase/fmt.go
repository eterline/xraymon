// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.

package usecase

import "strconv"

func PercentString(v float64) string {
	return strconv.Itoa(int(v)) + "%"
}

func CelsiusString(temp float64) string {
	if temp < (-273.15) {
		return "-273.15°C"
	}

	return strconv.FormatFloat(temp, 'f', 1, 64) + "°C"
}
