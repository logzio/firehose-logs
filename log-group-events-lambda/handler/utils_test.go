package handler

import (
	"github.com/stretchr/testify/assert"
	"os"
	"sort"
	"testing"
)

func TestValidateRequired(t *testing.T) {
	/* Missing all 3 required env variable */
	result := validateRequired()

	if result != nil {
		assert.Equal(t, "destination ARN must be set", result.Error())
	} else {
		t.Fatal("Expected an error, got nil")
	}

	err := os.Setenv(envFirehoseArn, "test-arn")
	if err != nil {
		return
	}

	/* Missing 2 required env variable */
	result = validateRequired()

	if result != nil {
		assert.Equal(t, "account id must be set", result.Error())
	} else {
		t.Fatal("Expected an error, got nil")
	}

	err = os.Setenv(envAccountId, "aws-account-id")
	if err != nil {
		return
	}

	/* Missing 1 required env variable */
	result = validateRequired()
	if result != nil {
		assert.Equal(t, "aws partition must be set", result.Error())
	} else {
		t.Fatal("Expected an error, got nil")
	}

	/* Valid required env variables */
	err = os.Setenv(envAwsPartition, "test-partition")
	if err != nil {
		return
	}

	result = validateRequired()
	assert.Nil(t, result)
}

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
