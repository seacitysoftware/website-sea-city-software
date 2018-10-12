package services

import (
	"fmt"
	"github.com/adbourne/website-seacitysoftware/domain"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/pkg/errors"
)

// ContactFormService is a service concerned with contact forms
type ContactFormService interface {
	// Process proceses the submitted contact form
	Process(contactForm *domain.ContactForm) error
}

// emailTemplate is the printf template for emails
const emailTemplate = `
Name: %s
Email: %s
Company: %s
Contact Number: %s
Message: %s
`
const AwsSesRejectedMessageError = "aws ses rejected message"
const AwsSesMailFromDomainNotVerifiedError = "aws ses mail from domain not verified"
const AwsSesConfigurationSetDoesNotExistError = "aws ses configuration set does not exist"
const AwsSesUnknownError = "aws ses unknown error"

// ContactFormEmailService is an implementation fo the ContactFormService which emails the submitted contact form
// to the provided email address
type ContactFormEmailService struct {
	Logger Logger

	EmailConfig *domain.EmailConfig

	// SesClient is the AWS Simple Email Service (SES) client
	SesClient *ses.SES
}

func (service *ContactFormEmailService) Process(contactForm *domain.ContactForm) (err error) {

	sesClient := service.SesClient
	sender := service.EmailConfig.Sender
	recipient := service.EmailConfig.Recipient
	subject := service.EmailConfig.Subject
	charSet := "UTF-8"

	textBody := contactFormToEmailBody(contactForm)

	sendEmailInput := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(textBody),
				},
				Text: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
	}

	// Attempt to send the email.
	result, err := sesClient.SendEmail(sendEmailInput)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				service.logSendEmailError(textBody, ses.ErrCodeMessageRejected, aerr.Error())
				err = errors.New(AwsSesRejectedMessageError)
				return

			case ses.ErrCodeMailFromDomainNotVerifiedException:
				service.logSendEmailError(textBody, ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
				err = errors.New(AwsSesMailFromDomainNotVerifiedError)
				return

			case ses.ErrCodeConfigurationSetDoesNotExistException:
				service.logSendEmailError(textBody, ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
				err = errors.New(AwsSesConfigurationSetDoesNotExistError)
				return

			default:
				service.logSendEmailError(textBody, "UNKNOWN", aerr.Error())
				err = errors.New(AwsSesUnknownError)
				return

			}
		} else {
			service.logSendEmailError(textBody, "UNKNOWN", err.Error())
			err = errors.New(AwsSesUnknownError)
			return
		}

		return
	}

	if nil != result.MessageId {
		service.Logger.Info(fmt.Sprintf("Contact form email sent with id '%s'", *result.MessageId), Fields{})
	} else {
		service.Logger.Info("Contact form email sent with but no message ID was provided", Fields{})
	}

	return
}

func (service *ContactFormEmailService) logSendEmailError(textBody string, reason string, error string) {
	service.Logger.Error("unable to send email", Fields{
		"email":  textBody,
		"reason": reason,
		"error":  error,
	})
}

func contactFormToEmailBody(contactForm *domain.ContactForm) string {
	return fmt.Sprintf(emailTemplate,
		contactForm.Name,
		contactForm.Email,
		contactForm.Company,
		contactForm.Number,
		contactForm.Message,
	)
}

// NewContactFormEmailService creates a new ContactFormEmailService
func NewContactFormEmailService(logger Logger, emailConfig *domain.EmailConfig, sesClient *ses.SES) ContactFormService {
	return &ContactFormEmailService{
		Logger:      logger,
		EmailConfig: emailConfig,
		SesClient:   sesClient,
	}
}
