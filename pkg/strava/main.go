package strava

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bdronneau/istravayou/pkg/models"
	"github.com/bdronneau/istravayou/pkg/utils"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"

	stravaSDK "github.com/bdronneau/go.strava"
)

// App of package
type App interface {
	NewHTTP(string)
}

type app struct {
	id     int
	secret string

	httpCORS string
	httpPort int

	authenticator *stravaSDK.OAuthAuthenticator
}

// Config for flags
type Config struct {
	id          *int
	secret      *string
	callbackURL *string

	dbHost     *string
	dbUser     *string
	dbName     *string
	dbPassword *string
	dbPort     *int

	httpCORS *string
	httpPort *int
}

type tokenURL struct {
	Code  string `json:"code"`
	State string
	Scope string
}

// Flags defines cli args for Strava Client
func Flags(fs *flag.FlagSet) Config {
	return Config{
		id:          fs.Int("app-id", 0, "[app] id of your application"),
		secret:      fs.String("app-secret", "", "[app] secret of your application"),
		callbackURL: fs.String("app-callback-url", "http://localhost:1323/private/exchange_token", "[app] URL callback by strava oauth"),
		httpCORS:    fs.String("http-cors", "http://localhost:8081", "[http] CORS domain"),
		httpPort:    fs.Int("http-port", 1323, "[http] api port"),

		dbHost:     fs.String("db-host", "localhost", "DB Host"),
		dbUser:     fs.String("db-user", "istravayou", "DB User"),
		dbName:     fs.String("db-name", "istravayou", "DB Name"),
		dbPassword: fs.String("db-password", "xoxo", "DB Password"),
		dbPort:     fs.Int("db-port", 5432, "DB Port"),
	}
}

// New allow to create an app
func New(config Config) (App, error) {
	authenticator := &stravaSDK.OAuthAuthenticator{
		CallbackURL:            *config.callbackURL,
		RequestClientGenerator: nil,
	}

	models.InitDB(fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", *config.dbUser, *config.dbPassword, *config.dbHost, *config.dbPort, *config.dbName))

	stravaSDK.ClientId = *config.id
	stravaSDK.ClientSecret = *config.secret

	return &app{
		id:            *config.id,
		secret:        *config.secret,
		authenticator: authenticator,

		httpCORS: *config.httpCORS,
		httpPort: *config.httpPort,
	}, nil
}

func (a app) NewHTTP(env string) {
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{a.httpCORS},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete},
	}))

	g := e.Group("/api")

	e.Use(middlewareLogger())
	g.Use(middlewareHeaders)

	// TODO delete me
	// Routes for internal testing
	if env == "dev" {
		e.GET("/private/login", a.handleLogin)
		e.GET("/private/exchange_token", echo.WrapHandler(a.authenticator.HandlerFunc(handleoAuthSuccess, handleoAuthFailure)))
		e.GET("/private/info", a.handleInfo)
	}

	// TODO: ofc improve this
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "It's okay, but not tested!")
	})

	// Routes for Front
	e.GET("/api/athlete", a.handleAthlete)
	e.HEAD("/api/athlete", a.handleHeadAthlete)
	e.POST("/api/auth", a.handleAuth)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", a.httpPort)))
}

// TODO: Should go in helpers
func middlewareLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			var err error
			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			reqSize := req.Header.Get(echo.HeaderContentLength)
			if reqSize == "" {
				reqSize = "0"
			}

			logrus.Infof("%s %s [%v] %s %-7s %s %3d %s %s %13v %s %s",
				id,
				c.RealIP(),
				stop.Format(time.RFC3339),
				req.Host,
				req.Method,
				req.RequestURI,
				res.Status,
				reqSize,
				strconv.FormatInt(res.Size, 10),
				stop.Sub(start).String(),
				req.Referer(),
				req.UserAgent(),
			)
			return err
		}
	}
}

func middlewareHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := utils.GetHeaderValue(c.Request().Header, "X-Athlete-Code")

		if err != nil {
			return c.JSON(400, "X-Athlete-Code is invalid")
		}
		return next(c)
	}
}
