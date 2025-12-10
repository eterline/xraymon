// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.

package usecase_test

import (
	"reflect"
	"testing"

	"github.com/eterline/xraymon/internal/utils/usecase"
)

func Test_RemoveSliceRepeats(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"empty slice", []string{}, []string{}},
		{"no repeats", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"with repeats", []string{"a", "b", "b", "c", "c", "c"}, []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := make([]string, len(tt.input))
			copy(s, tt.input)

			usecase.RemoveSliceRepeats(&s)

			if !reflect.DeepEqual(s, tt.expected) {
				t.Errorf("got %v, want %v", s, tt.expected)
			}
		})
	}
}

func Test_SliceWithoutRepeats(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{"empty slice", []int{}, []int{}},
		{"no repeats", []int{1, 2, 3}, []int{1, 2, 3}},
		{"with repeats", []int{1, 2, 2, 3, 3, 3}, []int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := usecase.SliceWithoutRepeats(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}

			if !reflect.DeepEqual(tt.input, tt.input) {
				t.Errorf("input slice was modified")
			}
		})
	}
}

func Test_SliceRepeats(t *testing.T) {
	slc := []string{"a", "b", "b", "c", "c", "c"}
	rpts := usecase.SliceRepeats(slc)

	expected := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	for k, v := range expected {
		if rpts[k] != v {
			t.Errorf("incorrect count for '%s': got %d, want %d", k, rpts[k], v)
		}
	}
}
