package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	middleware "github.com/wibecoderr/ToDo"
	"github.com/wibecoderr/ToDo/database"

	"github.com/wibecoderr/ToDo/handler"
)

func main() {
	if err := database.ConnectAndMigrate(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		database.SSLModeDisable); err != nil {
		logrus.Panicf("Failed to initialize and migrate database with error: %+v", err)
	}

	fmt.Println("Successfully migrated DB")
	r := chi.NewRouter()

	r.Post("/register", handler.RegisterUser)
	r.Post("/login", handler.LoginUser)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Post("/logout", handler.Logout)

		// remove s from todos route
		r.Post("/todo", handler.CreateTodo)
		r.Get("/todo", handler.GetAllTodos)

		r.Patch("/todo/{todoID}", handler.UpdateTodo)
		r.Delete("/todo/{id}", handler.DeleteTodo)
		r.Delete("/delete", handler.DeleteUser) //checked
	})

	http.ListenAndServe(":8080", r)

}
