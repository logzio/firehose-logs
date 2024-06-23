package common

import (
	"github.com/stretchr/testify/assert"
	"os"
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
	err := os.Setenv(envCustomGroups, "rand, custom")
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
