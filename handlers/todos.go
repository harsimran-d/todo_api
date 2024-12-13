package handlers

import (
	"context"
	"errors"
	"log"
	"time"
	"todo_api/types"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoHandler struct {
	DB *pgxpool.Pool
}

func NewTodoHandler(db *pgxpool.Pool) *TodoHandler {
	return &TodoHandler{
		DB: db,
	}
}

func (h *TodoHandler) GetAllTodos(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized access")
	}
	rows, err := h.DB.Query(context.Background(), "SELECT * FROM todos WHERE user_id=$1", userID)
	if err != nil {
		log.Println(err)
		return fiber.ErrInternalServerError
	}
	defer rows.Close()

	todos := []types.Todo{}
	for rows.Next() {
		var todo types.Todo
		if err := todo.Scan(rows); err != nil {
			log.Printf("Reading todo rows failed: %v", err)
			return fiber.ErrInternalServerError
		} else {
			todos = append(todos, todo)
		}
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return fiber.ErrInternalServerError
	}
	c.JSON(todos)
	return nil
}
func (h *TodoHandler) GetTodoById(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized access")
	}
	id := c.Params("id")
	row := h.DB.QueryRow(context.Background(), "SELECT * FROM todos WHERE id=$1 AND user_id=$2", id, userID)

	var todo types.Todo
	if err := todo.Scan(row); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Println(err)
			return fiber.ErrNotFound
		}
		log.Println(err)
		return fiber.ErrBadRequest
	}
	c.Status(fiber.StatusOK).JSON(todo)
	return nil
}

func (h *TodoHandler) PostTodo(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized access")
	}
	var createTodo struct {
		Title string
	}
	err := c.BodyParser(&createTodo)
	if err != nil {
		log.Println(err)
		return fiber.ErrBadRequest
	}
	if createTodo.Title == "" {
		log.Println("todo title was empty")
		return fiber.ErrBadRequest
	}
	result, err := h.DB.Exec(context.Background(), "INSERT INTO todos (title,user_id) VALUES ($1,$2)", createTodo.Title, userID)

	if err != nil {
		log.Println(err)
		return fiber.ErrInternalServerError
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected != 1 {
		log.Println("number of row affected not 1 while posting todo")
		return fiber.ErrInternalServerError
	}
	c.Status(fiber.StatusCreated)
	return nil
}
func (h *TodoHandler) UpdateTodo(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized access")
	}
	id := c.Params("id")
	var updateTodo struct {
		Title  *string
		Status *bool
	}
	err := c.BodyParser(&updateTodo)
	if err != nil {
		log.Println(err)
		return fiber.ErrBadRequest
	}
	if updateTodo.Title == nil && updateTodo.Status == nil {
		log.Println("Title and Status are empty")
		return fiber.NewError(fiber.StatusBadRequest, "Title or status is required")
	}
	if updateTodo.Title != nil && *updateTodo.Title == "" {
		log.Println("Title is empty")
		return fiber.NewError(fiber.StatusBadRequest, "Title cannot be empty")
	}

	row := h.DB.QueryRow(context.Background(), "UPDATE todos SET title=$1,status=$2,updated_at=$3 WHERE id=$4 AND user_id=$5 RETURNING id, title, status, created_at, updated_at,user_id", updateTodo.Title, updateTodo.Status, time.Now(), id, userID)
	var todo types.Todo
	err = todo.Scan(row)
	if err != nil {
		log.Println(err)
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Todo not found")
		}
		return fiber.ErrInternalServerError
	}
	c.JSON(todo)

	return nil
}

func (h *TodoHandler) DeleteTodo(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized access")
	}
	id := c.Params("id")
	result, err := h.DB.Exec(context.Background(), "DELETE FROM TODOS WHERE id=$1 AND user_id=$2", id, userID)
	if err != nil {
		log.Println(err)
		return fiber.ErrBadRequest
	}
	if count := result.RowsAffected(); count == 0 {
		log.Println("todo not found")
		return fiber.ErrNotFound
	}
	c.Status(fiber.StatusNoContent)
	return nil
}
