package services

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"time"
)

const (
	recaptchaServiceURL             = "https://www.google.com/recaptcha/api/siteverify"
	CannotCommunicateRecaptchaError = "unable to communicate with the recaptcha service"
	NotVerifiedError                = "user not verified by recaptcha"
)

type RecaptchaService interface {
	Verify(response string) error
}

type DefaultRecaptchaService struct {
	Logger Logger
	// Secret is the recaptcha secret
	Secret string

	HttpClient *http.Client
}

func (rs *DefaultRecaptchaService) Verify(response string) error {
	escapedSecret := url.QueryEscape(rs.Secret)
	escapedResponse := url.QueryEscape(response)

	requestURL := fmt.Sprintf("%s?secret=%s&response=%s", recaptchaServiceURL, escapedSecret, escapedResponse)

	request, err := http.NewRequest("POST", requestURL, nil)
	if err != nil {
		return err
	}

	var rawResponse *http.Response
	rawResponse, err = rs.HttpClient.Do(request)
	if err != nil {
		rs.Logger.Debug("Unable to communicate with Recaptcha verify service", Fields{"url": recaptchaServiceURL, "error": err.Error()})
		err = errors.New(CannotCommunicateRecaptchaError)
		return err
	}

	recaptchaResponse := NewBlankRecaptchaResponse()
	decoder := json.NewDecoder(rawResponse.Body)
	err = decoder.Decode(recaptchaResponse)
	if err != nil {
		rs.Logger.Debug("Recaptcha service responded with invalid JSON", Fields{"url": recaptchaServiceURL, "error": err.Error()})
		err = errors.New(CannotCommunicateRecaptchaError)
		return err
	}

	if !recaptchaResponse.Success {
		return errors.New(NotVerifiedError)
	}

	return nil
}

func NewDefaultRecaptchaService(secret string, logger Logger, httpClient *http.Client) *DefaultRecaptchaService {
	return &DefaultRecaptchaService{
		Logger:     logger,
		Secret:     secret,
		HttpClient: httpClient,
	}
}

type RecaptchaResponse struct {
	Success            bool      `json:"success"`
	ChallengeTimestamp time.Time `json:"challenge_ts"`
	HostName           string    `json:"hostname"`
	ErrorCodes         []string  `json:"error-codes"`
}

func NewBlankRecaptchaResponse() *RecaptchaResponse {
	return &RecaptchaResponse{
		Success:            false,
		ChallengeTimestamp: time.Time{},
		HostName:           "",
		ErrorCodes:         make([]string, 0),
	}
}
