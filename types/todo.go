package types

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Todo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    int       `json:"userId"`
}

func (t *Todo) String() string {
	return fmt.Sprintf("%s %s %t %s %s %d", t.ID, t.Title, t.Status, t.CreatedAt, t.UpdatedAt, t.UserID)
}

func (t *Todo) Scan(row pgx.Row) error {
	return row.Scan(&t.ID, &t.Title, &t.Status, &t.CreatedAt, &t.UpdatedAt, &t.UserID)
}
