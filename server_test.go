package main

import (
	"errors"
	"fmt"
	"github.com/adbourne/website-seacitysoftware/domain"
	"github.com/adbourne/website-seacitysoftware/mocks"
	"github.com/adbourne/website-seacitysoftware/services"
	"github.com/gin-gonic/gin/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"
)

const TemplatesDir = "./target/build/"

type ApplicationTestSuite struct {
	suite.Suite
	Logger services.Logger
	Port   int

	MockContactFormService *mocks.ContactFormService
	MockRecaptchaService   *mocks.RecaptchaService
	AppContext             *AppContext
}

func (suite *ApplicationTestSuite) SetupTest() {
	suite.Logger = newLogger()
	suite.Port = suite.getFreePort()
	suite.MockContactFormService = new(mocks.ContactFormService)
	suite.MockRecaptchaService = new(mocks.RecaptchaService)
	// TODO: Figure out why this can't be done in the test function
	suite.MockContactFormService.On("Process", mock.Anything).Return(nil)
	suite.MockRecaptchaService.On("Verify", mock.Anything).Return(nil)
	suite.AppContext = suite.createTestAppContext(suite.Port, suite.MockContactFormService, suite.MockRecaptchaService)
}

func (suite *ApplicationTestSuite) TestThatAFaviconExists() {
	go RunApp(suite.AppContext)
	homePageURL := fmt.Sprintf("http://localhost:%d/favicon.ico", suite.Port)
	suite.assertPageReturns200(homePageURL)
}

func (suite *ApplicationTestSuite) TestThatAHomePageExists() {
	go RunApp(suite.AppContext)

	homePageURL := fmt.Sprintf("http://localhost:%d/", suite.Port)
	suite.assertPageReturns200(homePageURL)
}

func (suite *ApplicationTestSuite) TestThatAContactPageExists() {
	go RunApp(suite.AppContext)

	contactPageURL := fmt.Sprintf("http://localhost:%d/contact", suite.Port)
	suite.assertPageReturns200(contactPageURL)
}

func (suite *ApplicationTestSuite) TestAContactPageCanBeSubmitted() {
	go RunApp(suite.AppContext)

	contactApiURL := fmt.Sprintf("http://localhost:%d/contact", suite.Port)

	contactForm := &domain.ContactForm{
		Name:              "Bob",
		Email:             "bob@someemail.com",
		Company:           "Bobcorp",
		Number:            "12345678",
		Message:           "Hey there!",
		RecaptchaResponse: "",
	}

	jsonBody, err := json.Marshal(contactForm)
	assert.NoError(suite.T(), err, "Unable to marshal contact form request")

	// Always process successfully
	//suite.MockContactFormService.On("Process", mock.Anything).Return(nil)
	//suite.AppContext.ContactFormService = suite.MockContactFormService

	err = Eventually(func() error {
		request, err := http.NewRequest("POST", contactApiURL, strings.NewReader(string(jsonBody)))
		assert.NoError(suite.T(), err, "unable to construct post request")
		request.Header.Set("Content-Type", "application/json")

		cookieJar, err := cookiejar.New(nil)
		assert.NoError(suite.T(), err, "Unable to construct cookie jar")
		assert.NoError(suite.T(), err, "Unable to construct URL")

		fmt.Printf("%+v\n", request)

		client := &http.Client{
			Timeout: 5 * time.Second,
			Jar:     cookieJar,
		}

		resp, err := client.Do(request)
		if err != nil {
			return err
		}

		actualStatusCode := resp.StatusCode
		if 201 != actualStatusCode {
			log.Printf("Expected expectedStatusCode code %d, but code status code %d", 201, actualStatusCode)
			return errors.New("received non 403 http status code")
		}

		return nil

	}, 10, 2*time.Second)

	assert.NoError(suite.T(), err, "201 not returned by contact form submission API")
	suite.MockContactFormService.AssertExpectations(suite.T())
}

func (suite *ApplicationTestSuite) TestThatAPrivacyPolicyPageExists() {
	go RunApp(suite.AppContext)

	privacyPageURL := fmt.Sprintf("http://localhost:%d/privacy", suite.Port)
	suite.assertPageReturns200(privacyPageURL)
}

func (suite *ApplicationTestSuite) TestThatACookiesPolicyPageExists() {
	go RunApp(suite.AppContext)

	privacyPageURL := fmt.Sprintf("http://localhost:%d/cookies", suite.Port)
	suite.assertPageReturns200(privacyPageURL)
}

func TestRunApplicationTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationTestSuite))
}

func (suite *ApplicationTestSuite) getFreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		assert.Fail(suite.T(), err.Error())
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		assert.Fail(suite.T(), err.Error())
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func (suite *ApplicationTestSuite) createTestAppContext(port int, contactFormService services.ContactFormService, recaptchaService services.RecaptchaService) *AppContext {
	logger := newLogger()

	pathToFrontend := getAbsolutePathOrPanic(TemplatesDir, logger)

	emailConfig := &domain.EmailConfig{
		Sender:          "",
		Recipient:       "",
		Subject:         "",
		AwsSesRegion:    "",
		AwsSesAccessKey: "",
		AwsSesSecretKey: "",
	}

	return &AppContext{
		Config: &domain.AppConfig{
			HttpPort:    port,
			EmailConfig: emailConfig,
		},
		TemplateDir:        pathToFrontend,
		Logger:             logger,
		ContactFormService: contactFormService,
		RecaptchaService:   recaptchaService,
	}
}

func (suite *ApplicationTestSuite) assertPageReturns200(url string) {

	suite.assertPageHasStatusCallback(url, 200, func(_ *http.Response) error {
		// Do nothing
		return nil
	})
}

func (suite *ApplicationTestSuite) assertPageHasStatusCallback(url string, expectedStatusCode int, callback func(resp *http.Response) error) {
	err := Eventually(func() error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}

		actualStatusCode := resp.StatusCode
		if expectedStatusCode != actualStatusCode {
			log.Printf("Expected expectedStatusCode code %d, but code expectedStatusCode code %d", expectedStatusCode, actualStatusCode)
			return errors.New("received non 200 http expectedStatusCode code")
		}

		return callback(resp)

	}, 10, 2*time.Second)

	assert.NoError(suite.T(), err, fmt.Sprintf("Page '%s' not found", url))
}

func Eventually(fx func() error, maxAttempts int, wait time.Duration) error {
	count := 0

	for count < maxAttempts {
		err := fx()
		if err == nil {
			return nil
		}

		count++
		time.Sleep(wait)
	}
	return errors.New("max attempts exceeded")
}
