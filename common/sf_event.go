package common

import (
	"encoding/json"
	"fmt"
)

type ActionType string

const (
	AddSF    ActionType = "add"
	UpdateSF ActionType = "update"
	DeleteSF ActionType = "delete"
)

type RequestParameters struct {
	Action      ActionType
	NewServices string `json:"newServices,omitempty"`
	OldServices string `json:"oldServices,omitempty"`
	NewCustom   string `json:"newCustom,omitempty"`
	OldCustom   string `json:"oldCustom,omitempty"`
	NewIsSecret string `json:"newIsSecret,omitempty"`
	OldIsSecret string `json:"oldIsSecret,omitempty"`
}

type Detail struct {
	EventName         string            `json:"eventName"`
	RequestParameters RequestParameters `json:"requestParameters"`
}

type SubscriptionFilterEvent struct {
	Detail Detail `json:"detail"`
}

func NewSubscriptionFilterEvent(params RequestParameters) SubscriptionFilterEvent {
	return SubscriptionFilterEvent{
		Detail: Detail{
			EventName:         "SubscriptionFilterEvent",
			RequestParameters: params,
		},
	}
}

func ConvertToRequestParameters(obj interface{}) (RequestParameters, error) {
	var rp RequestParameters
	bytes, err := json.Marshal(obj)
	if err != nil {
		return rp, fmt.Errorf("error marshalling interface: %v", err)
	}
	err = json.Unmarshal(bytes, &rp)
	if err != nil {
		return rp, fmt.Errorf("error unmarshalling to RequestParameters: %v", err)
	}
	return rp, nil
}
