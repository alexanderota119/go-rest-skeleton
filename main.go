package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go-rest-skeleton/infrastructure/authorization"
	"go-rest-skeleton/infrastructure/persistence"
	"go-rest-skeleton/interfaces"
	user_v1_0_0 "go-rest-skeleton/interfaces/handler/v1.0/user"
	welcome_v1_0_0 "go-rest-skeleton/interfaces/handler/v1.0/welcome"
	welcome_v2_0_0 "go-rest-skeleton/interfaces/handler/v2.0/welcome"
	"go-rest-skeleton/interfaces/middleware"
	"golang.org/x/exp/errors"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

var (
	rxURL = regexp.MustCompile(`^/regexp\d*`)
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic("no .env file provided")
	}
}

func main() {
	// Connect to DB: postgres | mysql
	dbDriver := os.Getenv("DB_DRIVER")
	dbHost := os.Getenv("DB_HOST")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbServices, errDb := persistence.NewRepositories(dbDriver, dbUser, dbPassword, dbHost, dbName, dbPort)
	if errDb != nil {
		panic(errDb)
	}
	defer dbServices.Close()
	dbServices.AutoMigrate()
	dbServices.Seeds()

	// Connect to redis
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisServices, errRedis := authorization.NewRedisDB(redisHost, redisPort, redisPassword)
	if errRedis != nil {
		panic(errRedis)
	}

	// Get options
	optSetLanguage := os.Getenv("APP_LANG")
	optSetDebug := os.Getenv("APP_ENV") != "production"
	optSetRequestID, _ := strconv.ParseBool(os.Getenv("ENABLE_REQUEST_ID"))
	optSetLogger, _ := strconv.ParseBool(os.Getenv("ENABLE_LOGGER"))
	optSetCors, _ := strconv.ParseBool(os.Getenv("ENABLE_CORS"))

	// Init authorization
	authToken := authorization.NewToken()
	authenticate := interfaces.NewAuthenticate(dbServices.User, redisServices.Auth, authToken)

	// Init interfaces
	welcomeApp := interfaces.NewWelcomeHandler(dbServices.User, authToken)
	welcomeV1 := welcome_v1_0_0.NewWelcomeHandler()
	welcomeV2 := welcome_v2_0_0.NewWelcomeHandler()
	userV1 := user_v1_0_0.NewUsers(dbServices.User, redisServices.Auth, authToken)

	// Logging
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: false,
		},
	)

	// Init gin with middleware
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(middleware.New(middleware.ResponseOptions{DebugMode: optSetDebug, DefaultLanguage: optSetLanguage}).Handler())
	router.Use(middleware.SetRequestID(middleware.RequestIDOptions{AllowSetting: optSetRequestID}))
	router.Use(middleware.CORSMiddleware(middleware.CORSOptions{AllowSetting: optSetCors}))
	router.Use(middleware.SetLogger(middleware.LoggerOptions{AllowSetting: optSetLogger}))
	router.Use(middleware.ApiVersion())
	v1 := router.Group("/api/v1/external")
	v2 := router.Group("/api/v2/external")

	// Routes
	// authorization
	v1.GET("/profile", middleware.AuthMiddleware(), authenticate.Profile)
	v1.POST("/login", authenticate.Login)
	v1.POST("/logout", middleware.AuthMiddleware(), authenticate.Logout)
	v1.POST("/refresh", authenticate.Refresh)
	v1.POST("/language", middleware.AuthMiddleware(), authenticate.SwitchLanguage)

	// users
	v1.GET("/users", middleware.AuthMiddleware(), userV1.GetUsers)
	v1.POST("/users", middleware.AuthMiddleware(), userV1.SaveUser)
	v1.GET("/users/:uuid", middleware.AuthMiddleware(), userV1.GetUser)

	// welcome
	v1.GET("/welcome_app", welcomeApp.Index)
	v1.GET("/welcome", welcomeV1.Index)
	v2.GET("/welcome", welcomeV2.Index)

	// ping
	v1.GET("/ping", func(c *gin.Context) {
		middleware.Formatter(c, nil, "pong")
	})

	// no route
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithError(http.StatusNotFound, errors.New("api_"))
	})

	// Run app at defined port
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8888"
	}
	panic(router.Run(":" + appPort))
}
