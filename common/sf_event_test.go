package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func stringPtr(s string) *string {
	/* helper function */
	return &s
}

func TestNewSubscriptionFilterEvent(t *testing.T) {
	tests := []struct {
		name     string
		params   RequestParameters
		expected SubscriptionFilterEvent
	}{
		{
			name: "Test New Add SF Event",
			params: RequestParameters{
				Action:      AddSF,
				NewServices: "newServices",
				NewCustom:   "newCustom",
				NewIsSecret: "false",
			},
			expected: SubscriptionFilterEvent{
				Detail: Detail{
					EventName: "SubscriptionFilterEvent",
					RequestParameters: RequestParameters{
						Action:      AddSF,
						NewServices: "newServices",
						NewCustom:   "newCustom",
						NewIsSecret: "false",
					},
				}},
		},
		{
			name: "Test New Update SF Event",
			params: RequestParameters{
				Action:      UpdateSF,
				NewServices: "newServices",
				OldServices: "oldServices",
				NewCustom:   "newCustom",
				OldCustom:   "someSecret",
				NewIsSecret: "false",
				OldIsSecret: "true",
			},
			expected: SubscriptionFilterEvent{
				Detail: Detail{
					EventName: "SubscriptionFilterEvent",
					RequestParameters: RequestParameters{
						Action:      UpdateSF,
						NewServices: "newServices",
						OldServices: "oldServices",
						NewCustom:   "newCustom",
						OldCustom:   "someSecret",
						NewIsSecret: "false",
						OldIsSecret: "true",
					},
				}},
		},
		{
			name: "Test New Delete SF Event",
			params: RequestParameters{
				Action:      DeleteSF,
				NewServices: "services",
				NewCustom:   "custom",
				NewIsSecret: "false",
			},
			expected: SubscriptionFilterEvent{
				Detail: Detail{
					EventName: "SubscriptionFilterEvent",
					RequestParameters: RequestParameters{
						Action:      DeleteSF,
						NewServices: "services",
						NewCustom:   "custom",
						NewIsSecret: "false",
					},
				}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := NewSubscriptionFilterEvent(test.params)
			assert.Equal(t, test.expected, result, "Expected %v, got %v", test.expected, result)
		})
	}
}

func TestConvertToRequestParameters(t *testing.T) {
	tests := []struct {
		name        string
		obj         interface{}
		expected    RequestParameters
		expectedErr *string
	}{
		{
			name: "Test Convert to Request Parameters",
			obj: map[string]interface{}{
				"action":       "add",
				"newServices":  "newServices",
				"newCustom":    "newCustom",
				"newIsSecret":  "false",
				"oldServices":  "oldServices",
				"oldCustom":    "oldCustom",
				"oldIsSecret":  "true",
				"invalidField": "invalid",
			},
			expected: RequestParameters{
				Action:      AddSF,
				NewServices: "newServices",
				NewCustom:   "newCustom",
				NewIsSecret: "false",
				OldServices: "oldServices",
				OldCustom:   "oldCustom",
				OldIsSecret: "true",
			},
			expectedErr: nil,
		},
		{
			name: "Test Convert to Request Parameters",
			obj: map[string]interface{}{
				"action": 123,
				"field":  "val",
			},
			expected:    RequestParameters{},
			expectedErr: stringPtr("error unmarshalling to RequestParameters: json: cannot unmarshal number into Go struct field RequestParameters.Action of type common.ActionType"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ConvertToRequestParameters(test.obj)

			if test.expectedErr == nil {
				assert.Equal(t, test.expected, result, "Expected %v, got %v", test.expected, result)
				assert.Nil(t, err, "Expected nil, got %v", err)
			} else {
				assert.Equal(t, *test.expectedErr, err.Error(), "Expected %v, got %v", test.expectedErr, err.Error())
			}
		})
	}
}
