package services

import (
	"fmt"
	"github.com/adbourne/website-seacitysoftware/domain"
	golog "log"
	"os"
	"strconv"
	"strings"
)

type ConfigService interface {
	// Loads the app config
	LoadConfig() *domain.AppConfig
}

const (
	// envVarHttpPort is the environment variable containing reference to the HTTP port to serve on
	envVarHttpPort = "HTTP_PORT"

	envVarFrontendDir = "FRONTEND_DIR"

	// envVarEmailSender environment variable containing reference to the email address to use in the "from" field
	envVarEmailSender = "EMAIL_SENDER"

	// envVarEmailRecipient to the email address to use in the "to" field
	envVarEmailRecipient = "EMAIL_RECIPIENT"

	envVarEmailSubject = "EMAIL_SUBJECT"

	envVarAwsSesRegion = "AWS_SES_REGION"

	enVarAwsSesAccessKey = "AWS_SES_ACCESS_KEY"

	enVarAwsSesSecretKey = "AWS_SES_SECRET_KEY"

	enVarRecaptchaSecret = "RECAPTCHA_SECRET"
)

type EnvVarConfigService struct {
	logger Logger
}

func NewEnvVarConfigService(logger Logger) *EnvVarConfigService {
	return &EnvVarConfigService{
		logger: logger,
	}
}

func (configService *EnvVarConfigService) LoadConfig() *domain.AppConfig {
	appConfig := &domain.AppConfig{
		HttpPort: configService.loadEnvVarAsIntOrPanic(envVarHttpPort),
		FrontendDir: configService.loadEnvVarAsStringOrPanic(envVarFrontendDir),
		EmailConfig: &domain.EmailConfig{
			Sender:          configService.loadEnvVarAsStringOrPanic(envVarEmailSender),
			Recipient:       configService.loadEnvVarAsStringOrPanic(envVarEmailRecipient),
			Subject:         configService.loadEnvVarAsStringOrPanic(envVarEmailSubject),
			AwsSesRegion:    configService.loadEnvVarAsStringOrPanic(envVarAwsSesRegion),
			AwsSesAccessKey: configService.loadEnvVarAsStringOrPanic(enVarAwsSesAccessKey),
			AwsSesSecretKey: configService.loadEnvVarAsStringOrPanic(enVarAwsSesSecretKey),
		},
		RecaptchaSecret: configService.loadEnvVarAsStringOrPanic(enVarRecaptchaSecret),
	}

	err := appConfig.Validate()
	if err != nil {
		panic(err.Error())
	}

	return appConfig
}

// LoadEnvVarAsIntOrPanic loadEnvVarAsInt is a utility function for loading an environment variable as an int and panics if it is not there
func (configService *EnvVarConfigService) loadEnvVarAsIntOrPanic(envVarName string) int {
	ev, isFound := os.LookupEnv(envVarName)
	if !isFound {
		panic(fmt.Sprintf("Environment variable '%s' was not found, application cannot start", envVarName))
	}

	evi, err := strconv.Atoi(ev)
	if err != nil {
		golog.Fatalf("Environment variable '%s' was not a number", envVarName)
		panic(fmt.Sprintf("Environment variable '%s' was not a number, application cannot start", envVarName))
	}

	configService.logger.Debug(fmt.Sprintf("Environment variable '%s' found as '%s'", envVarName, ev), make(Fields))
	return evi
}

// LoadEnvVarAsStringOrPanic loads an environment variable as a string and panics if it is not there
func (configService *EnvVarConfigService) loadEnvVarAsStringOrPanic(envVarName string) string {
	ev, isFound := os.LookupEnv(envVarName)
	if !isFound {
		msg := fmt.Sprintf("Environment variable '%s' not found", envVarName)
		panic(msg)
	}

	configService.logger.Debug(fmt.Sprintf("Environment variable '%s' found as '%s'", envVarName, ev), make(Fields))
	return ev
}

// splitBrokerList splits the comma delimited broker string into a slice of brokers
func splitBrokerList(brokers string) []string {
	return strings.Split(brokers, ",")
}
