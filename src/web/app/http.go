package main

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"html/template"
	"net/http"
	"strconv"
	"time"
	"web/common"
)

func RunHttp(services *Services) {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	e.Static("/static", "public/static")

	// Compile templates
	setRenderer := func() {
		e.Renderer = &Template{
			templates: template.Must(template.ParseGlob("public/views/*.html")),
		}
	}
	setRenderer()

	// Reload templates in development
	if services.env.Development {
		go func() {
			for {
				setRenderer()
				time.Sleep(time.Second)
			}
		}()
		e.Debug = true
	}

	// JWT
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(JwtCustomClaims)
		},
		SigningKey:  []byte(common.Sha256(services.env.TelegramKey)),
		TokenLookup: "cookie:token",
	}
	jwtConfigPublic := jwtConfig
	jwtConfigPublic.ContinueOnIgnoredError = true
	jwtConfigPublic.ErrorHandler = func(c echo.Context, err error) error {
		return nil
	}

	e.Use(servicesMiddleware(services))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = errorHandler

	// Public routes
	publicGroup := e.Group("")
	publicGroup.Use(echojwt.WithConfig(jwtConfigPublic))
	publicGroup.GET("/", viewLogin)
	publicGroup.POST("/login", loginDevice)
	publicGroup.POST("loginTelegram", loginTelegram)
	publicGroup.GET("/logout", logout)

	// Routes with authentication
	authGroup := e.Group("")
	authGroup.Use(echojwt.WithConfig(jwtConfig))
	authGroup.GET("/devices", viewDevices)
	authGroup.GET("/devices/:deviceName", viewDevice)
	authGroup.POST("/devices/:deviceName/users", addUser)
	authGroup.DELETE("/devices/:deviceName/users/:chatId", removeUser)
	authGroup.PUT("/devices/:deviceName/password", updatePassword)

	// Admin routes
	adminGroup := authGroup.Group("")
	adminGroup.Use(AdminMiddleware)
	adminGroup.GET("/admin", viewAdmin)
	adminGroup.POST("/devices", createDevice)
	adminGroup.DELETE("/devices/:deviceName", deleteDevice)

	e.Logger.Fatal(e.Start(":" + services.env.HttpPort))
}

func viewDevice(c echo.Context) error {
	requestDevice := c.Param("deviceName")
	if requestDevice == "" {
		return echo.ErrBadRequest
	}

	claims := getClaims(c)

	// Check if we are allowed through device login
	isAllowed := false
	if claims.Device != nil && claims.Device.DeviceName == requestDevice {
		isAllowed = true
	}

	// Check if we are allowed through telegram login
	if !isAllowed && claims.TelegramId != "" {
		devices, err := services.db.GetDevicesByUser(claims.TelegramId)
		if err != nil {
			c.Logger().Error(err)
			return echo.ErrInternalServerError
		}

		for _, device := range devices {
			if device.DeviceName == requestDevice {
				isAllowed = true
				break
			}
		}
	}

	if !isAllowed {
		return echo.ErrUnauthorized
	}

	// Get the users of the device
	users, err := services.db.GetDeviceUsers(requestDevice)
	if err != nil {
		c.Logger().Error(err)
		return echo.ErrInternalServerError
	}

	// Check if the user is an admin
	isAdmin, err := services.db.IsAdmin(requestDevice)
	if err != nil {
		c.Logger().Error(err)
		return echo.ErrInternalServerError
	}

	return c.Render(http.StatusOK, "dash", map[string]interface{}{
		"users":    users,
		"isAdmin":  isAdmin,
		"subtitle": requestDevice,
	})
}

func viewLogin(c echo.Context) error {
	return c.Render(http.StatusOK, "login", nil)
}

func viewDevices(c echo.Context) error {
	claims := getClaims(c)
	if claims.TelegramId == "" {
		return echo.ErrUnauthorized
	}

	devices, err := services.db.GetDevicesByUser(claims.TelegramId)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(500)
	}

	return c.Render(http.StatusOK, "devices", map[string]interface{}{
		"devices": devices,
	})
}

func addUser(c echo.Context) error {
	chatId := c.FormValue("chat_id")
	username := c.FormValue("username")
	deviceName := c.Param("deviceName")
	if username == "" || chatId == "" || deviceName == "" {
		return echo.ErrBadRequest
	}

	err := services.db.AddUser(deviceName, common.TelegramUser{
		ChatId: chatId,
		User:   username,
	})
	if err != nil {
		c.Logger().Error(err)
		return echo.ErrInternalServerError
	}

	return nil
}

func removeUser(c echo.Context) error {
	chatId := c.Param("chatId")
	if chatId == "" {
		return echo.ErrBadRequest
	}

	deviceName := c.Param("deviceName")
	if deviceName == "" {
		return echo.ErrBadRequest
	}

	err := services.db.DeleteUser(deviceName, chatId)
	if err != nil {
		c.Logger().Error(err)
		return echo.ErrInternalServerError
	}

	return nil
}

func updatePassword(c echo.Context) error {
	password := c.FormValue("password")
	passwordRepeat := c.FormValue("password-repeat")
	deviceName := c.Param("deviceName")
	if password != passwordRepeat || deviceName == "" {
		return echo.ErrBadRequest
	}

	err := services.db.UpdatePassword(deviceName, password)
	if err != nil {
		c.Logger().Error(err)
		return echo.ErrInternalServerError
	}

	return nil
}

func loginDevice(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	if username == "" || password == "" {
		return echo.ErrBadRequest
	}

	pass, err := services.db.Login(username, password)
	if err != nil {
		c.Logger().Error(err)
		return echo.ErrInternalServerError
	}

	if !pass {
		return echo.ErrUnauthorized
	}

	claims := getNewClaims(c, &JwtCustomClaims{
		Device: &common.Device{
			DeviceName: username,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	})
	err = setJwtCookie(c, claims)
	if err != nil {
		c.Logger().Error(err)
		return echo.ErrInternalServerError
	}

	return nil
}

func loginTelegram(c echo.Context) error {
	body := new(common.TelegramLogin)
	if err := c.Bind(body); err != nil {
		return echo.ErrBadRequest
	}

	// Check if the login is older than 24 hours
	if time.Now().Unix() > int64(body.AuthDate)+24*60*60 {
		return echo.ErrUnauthorized
	}

	// Check if the hash is valid
	if !body.Validate(services.env.TelegramKey) {
		return echo.ErrUnauthorized
	}

	// Create cookie of the full body
	username := body.Username
	if username == "" {
		username = body.FirstName + " " + body.LastName
	}

	claims := getNewClaims(c, &JwtCustomClaims{
		TelegramId:       strconv.Itoa(body.ID),
		TelegramUsername: username,
	})
	err := setJwtCookie(c, claims)
	if err != nil {
		c.Logger().Error(err)
		return echo.ErrInternalServerError
	}

	return nil
}

func logout(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:    "token",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
		Path:    "/",
	})

	return c.Redirect(http.StatusFound, "/")
}

func viewAdmin(c echo.Context) error {
	devices, err := services.db.GetDevices()
	if err != nil {
		c.Logger().Error(err)
		return echo.ErrInternalServerError
	}

	return c.Render(http.StatusOK, "admin", map[string]interface{}{
		"devices": devices,
	})
}

func createDevice(c echo.Context) error {
	username := c.FormValue("deviceName")
	password := c.FormValue("password")
	if username == "" || password == "" {
		return echo.ErrBadRequest
	}

	err := services.db.AddDevice(username, password)
	if err != nil {
		c.Logger().Error(err)
		return echo.ErrInternalServerError
	}

	return nil
}

func deleteDevice(c echo.Context) error {
	deviceName := c.Param("deviceName")
	if deviceName == "" || deviceName == "admin" {
		return echo.ErrBadRequest
	}

	err := services.db.DeleteDevice(deviceName)
	if err != nil {
		c.Logger().Error(err)
		return echo.ErrInternalServerError
	}

	return nil
}

func servicesMiddleware(services *Services) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("services", services)
			return next(c)
		}
	}
}

func errorHandler(err error, c echo.Context) {
	if err != nil {
		c.Logger().Error(err)
	}

	// if error is a custom HTTP error get the status code
	var he *echo.HTTPError
	if errors.As(err, &he) {
		err := c.Render(he.Code, "error", map[string]interface{}{
			"Error": err,
		})
		if err != nil {
			c.Logger().Error(err)
		}
		return
	}

	// if error is not a custom HTTP error return a 500
	err = c.Render(http.StatusInternalServerError, "error", map[string]interface{}{
		"Error": echo.ErrInternalServerError,
	})
	if err != nil {
		c.Logger().Error(err)
	}
}
