package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID    string `db:"uid" json:"id"`
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
	Age   int    `db:"age" json:"age"`
}

type Todo struct {
	ID          int        `db:"id" json:"id"`
	UserID      string     `db:"user_id" json:"user_id"`
	Title       string     `db:"title" json:"title"`
	Description string     `db:"description" json:"description"`
	Status      string     `db:"status" json:"status"`
	Deadline    *time.Time `db:"deadline" json:"deadline"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}
type Error struct {
	Error      string
	StatusCode int
	Message    string
}
type RegisterUserRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=6"`
	PhoneNumber string `json:"phone_number" validate:"required,min=9,max=10"`
	Age         int    `json:"age" validate:"required,min=1,max=120"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}
type UpdateTodoStatusRequest struct {
	Title       string `json:"title"  validate:"required,max=20"`
	Description string `json:"description" validate:"required,min=1"`
	Deadline    string `json:"deadline"`
}
type CreateTodoRequest struct {
	Title       string `json:"title" validate:"required,max=20"`
	Description string `json:"description" validate:"required,min=1"`
	Deadline    string `json:"deadline"`
}
type UserCxt struct {
	UserId    string `json:"user_id"`
	SessionId string `json:"session_id"`
}

type TodoRequest struct {
	Title       string `json:"title" validate:"required,max=20"`
	Description string `json:"description" validate:"required,min=1,max=200"`
	Status      string `json:"status"`
}
type Claims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}
type Result struct {
	UID      string `db:"uid"`
	Password string `db:"password"`
}
