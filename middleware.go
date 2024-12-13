package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleWare(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Missing or invalid token")
	}
	tokeString := authHeader[len("Bearer "):]
	secretKey := "JWT_SECRET"
	token, err := jwt.Parse(tokeString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired token")
	}
	sub, ok := claims["sub"].(float64)
	if !ok {
		fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired token")
	}
	c.Locals("userID", int(sub))

	return c.Next()
}
