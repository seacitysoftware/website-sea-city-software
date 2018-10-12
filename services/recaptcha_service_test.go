package services

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSuccessRecaptchaResponseCanBeUnMarshalled(t *testing.T) {
	actualResponse := NewBlankRecaptchaResponse()
	jsonResponse :=
		`
			{
			  "success": true,
			  "challenge_ts": "2018-08-08T13:32:16Z",
			  "hostname": "localhost"         	
			}
			`

	err := json.Unmarshal([]byte(jsonResponse), actualResponse)
	assert.NoError(t, err, "Unable to unmarshal recaptcha response")

	assert.Equal(t, true, actualResponse.Success)
	expectedTime, err := time.Parse(time.RFC3339, "2018-08-08T13:32:16Z")
	assert.NoError(t, err, "unable to parse expected time")
	assert.Equal(t, expectedTime, actualResponse.ChallengeTimestamp)
	assert.Equal(t, "localhost", actualResponse.HostName)
	assert.NotNil(t, actualResponse.ErrorCodes)
	assert.Equal(t, 0, len(actualResponse.ErrorCodes))
}

func TestErrorRecaptchaResponseCanBeUnMarshalled(t *testing.T) {
	actualResponse := NewBlankRecaptchaResponse()
	jsonResponse :=
		`
			{
			  "success": false,
			  "challenge_ts": "2018-08-08T13:32:16Z",
			  "hostname": "localhost",
			  "error-codes": [
              	"missing-input-secret",
				"missing-input-response"
              ]	
			}
			`

	err := json.Unmarshal([]byte(jsonResponse), actualResponse)
	assert.NoError(t, err, "Unable to unmarshal recaptcha response")

	assert.Equal(t, false, actualResponse.Success)
	expectedTime, err := time.Parse(time.RFC3339, "2018-08-08T13:32:16Z")
	assert.NoError(t, err, "unable to parse expected time")
	assert.Equal(t, expectedTime, actualResponse.ChallengeTimestamp)
	assert.Equal(t, "localhost", actualResponse.HostName)
	assert.NotNil(t, actualResponse.ErrorCodes)
	assert.Equal(t, 2, len(actualResponse.ErrorCodes))
	assert.Equal(t, "missing-input-secret", actualResponse.ErrorCodes[0])
	assert.Equal(t, "missing-input-response", actualResponse.ErrorCodes[1])
}
