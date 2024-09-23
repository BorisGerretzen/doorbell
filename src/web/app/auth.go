package main

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
	"web/common"
)

type JwtCustomClaims struct {
	TelegramId       string         `json:"telegram_id"`
	TelegramUsername string         `json:"telegram_username"`
	Device           *common.Device `json:"device"`
	jwt.RegisteredClaims
}

func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims := getClaims(c)

		if claims.Device != nil && claims.Device.Admin {
			return next(c)
		}

		if claims.TelegramId != "" {
			devices, err := services.db.GetDevicesByUser(claims.TelegramId)
			if err != nil {
				c.Logger().Error(err)
				return echo.NewHTTPError(500)
			}

			for _, device := range devices {
				if device.Admin {
					return next(c)
				}
			}
		}

		return echo.ErrUnauthorized
	}
}

func getNewClaims(c echo.Context, newClaims *JwtCustomClaims) JwtCustomClaims {
	claims := getClaims(c)

	if newClaims.TelegramId != "" {
		claims.TelegramId = newClaims.TelegramId
	}
	if newClaims.TelegramUsername != "" {
		claims.TelegramUsername = newClaims.TelegramUsername
	}
	if newClaims.Device != nil {
		claims.Device = newClaims.Device
	}
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
	}

	return claims
}

func getClaims(c echo.Context) JwtCustomClaims {
	if c.Get("user") == nil {
		return JwtCustomClaims{}
	}

	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*JwtCustomClaims)
	return *claims
}

func setJwtCookie(c echo.Context, claims JwtCustomClaims) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(common.Sha256(services.env.TelegramKey)))
	if err != nil {
		return err
	}
	c.SetCookie(&http.Cookie{
		Name:    "token",
		Value:   t,
		Expires: time.Now().Add(24 * time.Hour),
		Path:    "/",
	})
	return nil
}
