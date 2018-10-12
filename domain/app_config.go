package domain

import (
	"github.com/pkg/errors"
	"regexp"
)

const (
	EmailInvalidError        = "provided email address was not valid"
	HttpPortInvalidError     = "provided HTTP port was not valid"
	EmailSubjectInvalidError = "provided email subject was invalid"
	AwsSesInvalidRegionError = "provided AWS SES region is invalid"
)

// AppConfig is the application's configuration
type AppConfig struct {
	// HttpPort is the port to run on
	HttpPort int

	FrontendDir string

	// EmailConfig is the email configuration
	EmailConfig *EmailConfig

	RecaptchaSecret string
}

func (appConfig *AppConfig) Validate() (err error) {
	if appConfig.HttpPort == 0 {
		err = errors.New(HttpPortInvalidError)
		return
	}

	err = appConfig.EmailConfig.Validate()
	if err != nil {
		return
	}

	return
}

type EmailConfig struct {
	// Sender is the email address to put in the "from" field when sending emails
	Sender string

	// Recipient is the email address to send emails to
	Recipient string

	// Subject is the subject to use in the emails
	Subject string

	// AwsSesRegion is the AWS SES region to use when sending emails
	AwsSesRegion string

	// AwsSesAccessKey is the AWS SES IAM user access key
	AwsSesAccessKey string

	// AwsSesSecretKey is the AWS SES IAM user secret key
	AwsSesSecretKey string
}

func (emailConfig *EmailConfig) Validate() (err error) {
	// Validate the sender
	err = validateEmailFormat(emailConfig.Sender)
	if err != nil {
		return
	}

	// Validate the recipient
	err = validateEmailFormat(emailConfig.Recipient)
	if err != nil {
		return
	}

	// Validate the email subject
	if len(emailConfig.Subject) <= 0 {
		err = errors.New(EmailSubjectInvalidError)
		return
	}

	// Validate the AWS SES region
	if len(emailConfig.AwsSesRegion) <= 0 {
		err = errors.New(AwsSesInvalidRegionError)
		return
	}

	return
}

// validateEmailFormat validates that the provided email address is in the correct format
func validateEmailFormat(email string) (err error) {
	emailFormatRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if !emailFormatRegex.MatchString(email) {
		err = errors.New(EmailInvalidError)
	}

	return
}
