package handler

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestGetServicesMap(t *testing.T) {
	result := getServicesMap()
	assert.NotNil(t, result)
}

func TestConvertStrToArr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string(nil),
		},
		{
			name:     "single element",
			input:    "service1",
			expected: []string{"service1"},
		},
		{
			name:     "multiple elements",
			input:    "service1, service2",
			expected: []string{"service1", "service2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convertStrToArr(test.input)
			assert.Equal(t, test.expected, result, "Expected %v, got %v", test.expected, result)
		})
	}
}

func TestFindDifferences(t *testing.T) {
	tests := []struct {
		name             string
		old              []string
		new              []string
		expectedToAdd    []string
		expectedToRemove []string
	}{
		{
			name:             "no differences",
			old:              []string{"service1", "service2"},
			new:              []string{"service1", "service2"},
			expectedToAdd:    []string(nil),
			expectedToRemove: []string(nil),
		},
		{
			name:             "delete all",
			old:              []string{"service1", "service2"},
			new:              []string{},
			expectedToAdd:    []string(nil),
			expectedToRemove: []string{"service1", "service2"},
		},
		{
			name:             "add to empty",
			old:              []string{},
			new:              []string{"service1", "service2"},
			expectedToAdd:    []string{"service1", "service2"},
			expectedToRemove: []string(nil),
		},
		{
			name:             "delete some and add others",
			old:              []string{"service1", "service2"},
			new:              []string{"service1", "service3"},
			expectedToAdd:    []string{"service3"},
			expectedToRemove: []string{"service2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resultToAdd, resultToRemove := findDifferences(test.old, test.new)
			sort.Strings(resultToAdd)
			sort.Strings(resultToRemove)
			assert.Equal(t, test.expectedToAdd, resultToAdd, "Expected %v, got %v", test.expectedToAdd, resultToAdd)
			assert.Equal(t, test.expectedToRemove, resultToRemove, "Expected %v, got %v", test.expectedToRemove, resultToRemove)
		})
	}
}
