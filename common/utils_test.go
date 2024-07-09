package common

import (
	"github.com/stretchr/testify/assert"
	"os"
	"sort"
	"testing"
)

func TestGetServicesNoServices(t *testing.T) {
	result := GetServices()
	assert.Nil(t, result)
}

func TestGetServices(t *testing.T) {
	err := os.Setenv(envServices, "rds, cloudwatch, custom")
	if err != nil {
		return
	}

	result := GetServices()
	assert.Equal(t, []string{"rds", "cloudwatch", "custom"}, result)
}

func TestGetServicesMap(t *testing.T) {
	result := GetServicesMap()
	assert.NotNil(t, result)
}

func TestGetCustomPathsNoPaths(t *testing.T) {
	result := GetCustomPaths()
	assert.Nil(t, result)
}

func TestGetCustomPaths(t *testing.T) {
	err := os.Setenv(EnvCustomGroups, "rand, custom")
	if err != nil {
		return
	}

	result := GetCustomPaths()
	assert.Equal(t, []string{"rand", "custom"}, result)
}

func TestParseServices(t *testing.T) {
	tests := []struct {
		name        string
		servicesStr string
		expected    []string
	}{
		{
			name:        "no service",
			servicesStr: "",
			expected:    nil,
		},
		{
			name:        "single service",
			servicesStr: "service",
			expected:    []string{"service"},
		},
		{
			name:        "multiple services",
			servicesStr: "service, another, oneMore",
			expected:    []string{"service", "another", "oneMore"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ParseServices(test.servicesStr)
			assert.Equal(t, test.expected, result, "Expected %v, got %v", test.expected, result)
		})
	}
}

func TestListContains(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		lst      []string
		expected bool
	}{
		{
			name:     "empty list",
			str:      "some string",
			lst:      []string{},
			expected: false,
		},
		{
			name:     "string not in the list",
			str:      "item3",
			lst:      []string{"item1", "item2"},
			expected: false,
		},
		{
			name:     "string in the list",
			str:      "item1",
			lst:      []string{"item1", "item2"},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ListContains(test.str, test.lst)
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
			resultToAdd, resultToRemove := FindDifferences(test.old, test.new)
			sort.Strings(resultToAdd)
			sort.Strings(resultToRemove)
			assert.Equal(t, test.expectedToAdd, resultToAdd, "Expected %v, got %v", test.expectedToAdd, resultToAdd)
			assert.Equal(t, test.expectedToRemove, resultToRemove, "Expected %v, got %v", test.expectedToRemove, resultToRemove)
		})
	}
}

func TestGetSecretNameFromArn(t *testing.T) {
	err := os.Setenv(EnvAwsRegion, "us-east-1")
	if err != nil {
		return
	}
	err = os.Setenv(envAccountId, "486140753397")
	if err != nil {
		return
	}

	arn := "arn:aws:secretsmanager:us-east-1:486140753397:secret:testSecretName-56y7ud"
	assert.Equal(t, "testSecretName", GetSecretNameFromArn(arn))

	arn = "arn:aws:secretsmanager:us-east-1:486140753397:secret:random-name-56y7ud"
	assert.Equal(t, "random-name", GetSecretNameFromArn(arn))

	arn = "arn:aws:secretsmanager:us-east-1:486140753397:secret:now1with2numbers345-56y7ud"
	assert.Equal(t, "now1with2numbers345", GetSecretNameFromArn(arn))
}
