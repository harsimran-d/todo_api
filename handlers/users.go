package handlers

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UsersHandler struct {
	DB *pgxpool.Pool
}

func NewUserHandler(db *pgxpool.Pool) *UsersHandler {
	return &UsersHandler{
		DB: db,
	}
}

func (h *UsersHandler) SignIn(c *fiber.Ctx) error {
	var signinUser struct {
		Username *string
		Password *string
	}

	err := c.BodyParser(&signinUser)
	if err != nil || signinUser.Username == nil || signinUser.Password == nil {
		return fiber.NewError(fiber.StatusBadRequest, "enter Username and Password in body")
	}
	if *signinUser.Username == "" || *signinUser.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "values cannot be empty")
	}
	row := h.DB.QueryRow(context.Background(), "SELECT id,hashed_password from users where username=$1", *signinUser.Username)
	var user struct {
		Id             int
		HashedPassword string
	}
	err = row.Scan(&user.Id, &user.HashedPassword)
	if err != nil {
		log.Printf("Error parsing user info: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Error Signing in")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(*signinUser.Password))
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "check username or password")
	}
	var (
		key []byte
		t   *jwt.Token
		s   string
	)
	if keyString := "JWT_SECRET"; keyString == "" {
		log.Printf("could not load JWT_SECRET: %v", keyString)
		return fiber.NewError(fiber.StatusInternalServerError, "Error signing in")
	} else {
		key = []byte(keyString)
	}

	t = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "todo_server",
		"sub": user.Id,
	})
	s, err = t.SignedString(key)
	if err != nil {
		log.Printf("could not create jwt: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Error signing in")
	}
	c.JSON(map[string]string{"token": s})
	c.Status(fiber.StatusOK)
	return nil
}

func (h *UsersHandler) SignUp(c *fiber.Ctx) error {
	var signupUser struct {
		Name            *string
		Username        *string
		Password        *string
		ConfirmPassword *string
	}
	err := c.BodyParser(&signupUser)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}
	if signupUser.Username == nil || signupUser.Password == nil || signupUser.ConfirmPassword == nil {
		return fiber.NewError(fiber.StatusBadRequest, "missing need values")
	}
	if *signupUser.Username == "" || *signupUser.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "username or password cannot be empty")
	}
	if *signupUser.Password != *signupUser.ConfirmPassword {
		return fiber.NewError(fiber.StatusBadRequest, "passwords do not match")
	}
	if signupUser.Name == nil || *signupUser.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name cannot be empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*signupUser.Password), 12)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Something went wrong")
	}
	_, err = h.DB.Exec(context.Background(), "INSERT INTO users (name,username,hashed_password) VALUES ($1,$2,$3)", *signupUser.Name, *signupUser.Username, hashedPassword)
	if err != nil {
		log.Println(err.Error())
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fiber.NewError(fiber.StatusConflict, "username already taken")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "error creating user")
	}

	c.Status(fiber.StatusCreated)
	c.SendString("user created")
	return nil
}
