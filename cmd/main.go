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
		r.Delete("/delete", handler.DeleteUser)

		r.Post("/todos", handler.CreateTodo)
		r.Get("/todos", handler.GetAllTodos)
		//r.Get("/todos", handler.GetTodos)
		r.Get("/todos/session", handler.GetBysession)
		r.Get("/search", handler.SearchTodo)
		r.Get("/todos/status", handler.GetByStatus) // change status
		r.Get("/todos/{todoid}", handler.GetByTodoID)
		r.Get("/todos/filter", handler.UpcomingTodos) // not filter, upcomingTodo
		// some other routes: /completed, /incompleted

		r.Patch("/todos/{id}", handler.UpdateTodo)
		r.Delete("/todos/{id}", handler.DeleteTodo) //checked
	})
	//r.Get("/todos", handler.GetTodosHandler)    // checked  , title search , archieve at
	http.ListenAndServe(":8080", r)
	// indexing , routs , schema , session , filter-get todo, unique index , refrence , hashing
}
