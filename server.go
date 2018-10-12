package main

import (
	"fmt"
	"github.com/adbourne/website-seacitysoftware/domain"
	"github.com/adbourne/website-seacitysoftware/services"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/sirupsen/logrus"
	"net/http"
	"path/filepath"
	"time"
	"github.com/labstack/echo"
	"html/template"
	"io"
	"gopkg.in/go-playground/validator.v9"
	"github.com/unrolled/secure"
)

func main() {
	logger := newLogger()

	// Load the application config
	configService := newConfigService(logger)
	appConfig := configService.LoadConfig()
	emailConfig := appConfig.EmailConfig

	// Get the absolute path to the templates directory
	templateDir := appConfig.FrontendDir
	absTemplateDir := getAbsolutePathOrPanic(templateDir, logger)

	// Create the services
	sesClient := newSesClient(emailConfig)
	contactFormService := newContactFormService(logger, emailConfig, sesClient)
	httpClient := newHttpClient()
	recaptchaService := newRecaptchaService(appConfig.RecaptchaSecret, logger, httpClient)

	// Create the AppContext
	ctx := &AppContext{
		Config:             appConfig,
		TemplateDir:        absTemplateDir,
		Logger:             logger,
		ContactFormService: contactFormService,
		RecaptchaService:   recaptchaService,
		//ContactPageHandler:  contactPageHandler,
	}

	// Run the application
	RunApp(ctx)
}

// getAbsolutePathOrPanic gets the absolute path of the provide path or panics
func getAbsolutePathOrPanic(frontendDirectory string, logger services.Logger) string {
	absPath, err := filepath.Abs(frontendDirectory)
	if err != nil {
		logger.Error("Unable to get the absolute path for provide directory", services.Fields{"path": absPath})
		panic(err)
	}
	logger.Info("Found directory", services.Fields{"path": absPath})
	return absPath
}

// AppContext is the application's context
type AppContext struct {
	// Port is the port the application should run on
	Config *domain.AppConfig

	// TemplateDir is the directory containing the templates
	TemplateDir string

	// Logger is the application's logger
	Logger services.Logger

	// ContactFormService is a service responsible for dealing with Contact Forms
	ContactFormService services.ContactFormService

	// RecaptchaService is a service responsible for interacting with recpatcha
	RecaptchaService services.RecaptchaService

	// ContactPageHandler is the handler for the contact page
	ContactPageHandler http.Handler
}

func newLogger() services.Logger {
	logrusLogger := logrus.New()
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logger := services.NewLogrusLogger(logrusLogger)
	logger.Info("Starting logger", services.Fields{})
	return logger
}

func newHttpClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 15,
	}
}

func newConfigService(logger services.Logger) services.ConfigService {
	return services.NewEnvVarConfigService(logger)
}

func newSesClient(emailConfig *domain.EmailConfig) *ses.SES {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(emailConfig.AwsSesRegion),
		Credentials: credentials.NewStaticCredentials(emailConfig.AwsSesAccessKey, emailConfig.AwsSesSecretKey, ""),
	})
	if err != nil {
		panic(err.Error())
	}

	return ses.New(sess)
}

func newRecaptchaService(secret string, logger services.Logger, httpClient *http.Client) services.RecaptchaService {
	return services.NewDefaultRecaptchaService(secret, logger, httpClient)
}

func newContactFormService(logger services.Logger, emailConfig *domain.EmailConfig, sesClient *ses.SES) services.ContactFormService {
	return services.NewContactFormEmailService(logger, emailConfig, sesClient)
}

//func newContactPageHandler(contactFormFilePath string, logger services.Logger, csrfService services.CsrfService) http.Handler {
//	return handler.NewContactHandler(contactFormFilePath, logger, csrfService)
//}

type EchoTemplate struct {
	templates *template.Template
}

func (t *EchoTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type CustomEchoErrorHandler struct {
	Logger services.Logger
}

func (ceh *CustomEchoErrorHandler) handle(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	ceh.Logger.Error("Rendering error page", services.Fields{
		"error": err.Error(),
	})

	err = c.Render(code, fmt.Sprintf("%d.html", code), code)
	if err != nil {
		ceh.Logger.Error("Unable to render error page", services.Fields{
			"error": err.Error(),
		})
	}
}

func RunApp(ctx *AppContext) {
	logger := ctx.Logger

	e := echo.New()

	secureMiddleware := secure.New(secure.Options{
		FrameDeny:          true,
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
	})
	e.Use(echo.WrapMiddleware(secureMiddleware.Handler))

	// Configure echo logging
	//e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
	//	Format: "method=${method}, uri=${uri}, status=${status}\n",
	//}))

	// Configure echo error handling
	customerErrorHandler := &CustomEchoErrorHandler{
		Logger: logger,
	}
	e.HTTPErrorHandler = customerErrorHandler.handle

	// Configure templates
	viewsDir := filepath.Join(getAbsolutePathOrPanic("views", ctx.Logger), "*.html")
	logger.Info("Views directory", services.Fields{
		"viewsDir": viewsDir,
	})
	t := &EchoTemplate{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
	e.Renderer = t

	// Register the public dir
	e.Static("/", "public")

	// Register endpoints
	e.GET("/", func(c echo.Context) error {

		params := map[string]interface{}{
			"Tagline":        "We build great software and enable others to do the same",
			"TaglineSummary": "From concept to reality. We provide the services and the support to make your software project a success.",
		}

		return c.Render(http.StatusOK, "index.html", params)
	})

	e.GET("/services", func(c echo.Context) error {
		params := map[string]interface{}{
		}

		return c.Render(http.StatusOK, "index.html", params)
	})

	e.GET("/privacy", func(c echo.Context) error {
		params := map[string]interface{}{
			"Tagline":        "Privacy notice",
			"TaglineSummary": "This privacy notice is for visitors of this website.",
		}

		return c.Render(http.StatusOK, "privacy.html", params)
	})

	e.GET("/cookies", func(c echo.Context) error {
		params := map[string]interface{}{
			"Tagline":        "Cookies Policy",
			"TaglineSummary": "This cookie policy is for visitors of this website.",
		}

		return c.Render(http.StatusOK, "cookies.html", params)
	})

	e.GET("/contact", func(c echo.Context) error {
		params := map[string]interface{}{
			"Tagline":        "Contact",
			"TaglineSummary": "Have a question? Want to chat about a project you're working on? Fill in the form below or drop us an email. We'll be right with you.",
		}

		return c.Render(http.StatusOK, "contact.html", params)
	})

	e.POST("/contact", func(c echo.Context) error {

		contactFormSubmission := &domain.ContactForm{
			Name:              c.FormValue("name"),
			Email:             c.FormValue("email"),
			Company:           c.FormValue("company"),
			Number:            c.FormValue("number"),
			Message:           c.FormValue("message"),
			RecaptchaResponse: c.FormValue("g-recaptcha-response"),
		}

		validate := validator.New()
		err := validate.Struct(contactFormSubmission)
		if err != nil {
			logger.Error("Received an invalid contact form", services.Fields{"error": err.Error()})
			return c.JSON(400, "")
		}

		recaptchaService := ctx.RecaptchaService
		err = recaptchaService.Verify(contactFormSubmission.RecaptchaResponse)
		if err != nil {
			errorMessage := err.Error()
			if errorMessage == services.CannotCommunicateRecaptchaError {
				logger.Error("Unable to communicate with Recaptcha Service", services.Fields{"error": err.Error()})
				return c.JSON(500, "")
			} else if errorMessage == services.NotVerifiedError {
				logger.Warn("Received a submit contact form request for a user not verified by Recaptcha", services.Fields{"error": err.Error()})
				return c.JSON(403, "")
			} else {
				logger.Warn("Unexpected error verifying contact form submission with Recaptcha", services.Fields{"error": err.Error()})
				return c.JSON(500, "")
			}
		}

		err = ctx.ContactFormService.Process(contactFormSubmission)
		if err != nil {
			logger.Error("Unable to process contact form", services.Fields{"error": err.Error()})
			return c.JSON(500, "")
		}

		return c.JSON(200, "")

	})

	// TODO: move to service
	httpPort := ctx.Config.HttpPort
	logger.Info("Starting server", services.Fields{"port": httpPort})

	// TODO: Timeouts
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", httpPort)))
}
