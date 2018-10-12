package domain

type ContactForm struct {
	Name              string `json:"name" validate:"required"`
	Email             string `json:"email" validate:"required"`
	Company           string `json:"company" validate:"required"`
	Number            string `json:"number" validate:"required"`
	Message           string `json:"message" validate:"required"`
	RecaptchaResponse string `json:"recaptchaResponse" validate:"required"`
}
