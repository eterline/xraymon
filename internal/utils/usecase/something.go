// Copyright (c) 2025 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.

package usecase

import "time"

func MapSlicesLen[K comparable, V any](m map[K]V) (keys []K, values []V, n int) {
	mapLen := len(m)

	if mapLen == 0 {
		return nil, nil, 0
	}

	keys = make([]K, 0, mapLen)
	values = make([]V, 0, mapLen)

	for key, value := range m {
		keys = append(keys, key)
		values = append(values, value)
	}

	return keys, values, mapLen
}

func MapSlices[K comparable, V any](m map[K]V) (keys []K, values []V) {
	keys, values, _ = MapSlicesLen(m)
	return keys, values
}

func MsToDuration[T number](ms T) time.Duration {
	return time.Millisecond * time.Duration(ms)
}

func RemoveSliceRepeats[T comparable](s *[]T) {
	seen := map[T]struct{}{}
	idx := 0

	for _, v := range *s {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			(*s)[idx] = v
			idx++
		}
	}

	*s = (*s)[:idx]
}

func SliceWithoutRepeats[T comparable](s []T) []T {
	newS := make([]T, len(s))
	copy(newS, s)
	RemoveSliceRepeats(&newS)
	return newS
}

func SliceRepeats[T comparable](s []T) map[T]int {
	ct := map[T]int{}

	for _, v := range s {
		ct[v]++
	}

	return ct
}
