package main

import (
	"context"
	"log"
	"todo_api/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	app := fiber.New()
	dsn := "postgres://user:password@localhost:5432/database?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer pool.Close()
	todoHandlers := handlers.NewTodoHandler(pool)
	userHandlers := handlers.NewUserHandler(pool)
	app.Post("/signin", userHandlers.SignIn)
	app.Post("/signup", userHandlers.SignUp)
	app.Use(AuthMiddleWare)
	app.Get("/todos", todoHandlers.GetAllTodos)
	app.Get("/todos/:id", todoHandlers.GetTodoById)
	app.Post("/todos/", todoHandlers.PostTodo)
	app.Put("/todos/:id", todoHandlers.UpdateTodo)
	app.Delete("/todos/:id", todoHandlers.DeleteTodo)
	app.Listen(":3000")
}
