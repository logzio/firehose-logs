package common

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSession(t *testing.T) {
	result, err := GetSession()
	assert.IsType(t, (*session.Session)(nil), result)
	assert.Nil(t, err)
}
