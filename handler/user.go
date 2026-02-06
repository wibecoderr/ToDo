package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/wibecoderr/ToDo"
	models "github.com/wibecoderr/ToDo/Model"

	utils "github.com/wibecoderr/ToDo"
	database2 "github.com/wibecoderr/ToDo/database"
	database "github.com/wibecoderr/ToDo/database/dbhelper"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterUserRequest

	// parse body - utils
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse body")
		return
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		//
		utils.RespondValidationError(w, errs)

	}

	// Check if user already exists
	exists, err := database.IsUserRegistered(req.Email)
	if err != nil {
		// utils
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "user already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, err := database.CreateUser(
		req.Email,
		req.Name,
		hashedPassword,
		req.PhoneNumber,
		req.Age,
	)

	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create user")
		return
	}

	// Create session for auto-login
	sessionToken, err := database.CreateSession(database2.Todo, userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create session")
		return
	}

	// utils - Respond JSON
	utils.RespondJSON(w, http.StatusCreated, map[string]string{
		"user_id":       userID,
		"session_token": sessionToken,
		"message":       "registered and logged in",
	})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest

	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse body")
		return
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}
	userID, storedPassword, err := database.GetUserByEmail(req.Email)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "server error")
		return
	}
	// for delete
	if userID == "" {
		// check wheether user exist or not is exists tell user to contact
		archivedUserID, err := database.GetArchivedUserByEmail(req.Email)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "server error")
			return
		}

		if archivedUserID != "" {
			utils.RespondError(w, http.StatusBadRequest, err, "user_id is not registered")
			return
		}

		utils.RespondError(w, http.StatusBadRequest, err, "invalid user")
	}

	if !utils.CheckPasswordHash(req.Password, storedPassword) {
		utils.RespondError(w, http.StatusUnauthorized, err, "invalid credentials")
		return
	}

	sessionToken, err := database.CreateSession(database2.Todo, userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "server error")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, map[string]string{
		"session_token": sessionToken,
	})
}

func GetAllTodos(w http.ResponseWriter, r *http.Request) {
	SessionToken, err := utils.GetSessionToken(r)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, err, "unauthorized")
		return
	}

	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		if s != "active" && s != "completed" && s != "incomplete" {
			utils.RespondError(w, http.StatusBadRequest, err, "invalid status")
			return
		}
		status = &s
	}

	days := 0
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, err, "invalid days parameter: must be an integer")
			return
		}
		if d < 0 {
			utils.RespondError(w, http.StatusBadRequest, err, "invalid days parameter: must be non-negative")
			return
		}
		days = d
	}

	todos, err := database.GetFilteredTodos(SessionToken, status, days)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to retrieve todos")
		return
	}

	utils.RespondJSON(w, http.StatusOK, todos)
}

// put this model.go

func GetByTodoID(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "server error")
		return
	}

	todoID := chi.URLParam(r, "todoId")
	if todoID == "" {
		utils.RespondError(w, http.StatusBadRequest, err, "todo is not right")
		return
	}

	// pass user id also
	todo, err := database.GetTodoByID(todoID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "server error")
		return
	}

	if todo.UserID != userID {
		utils.RespondError(w, http.StatusUnauthorized, err, "user is not authorized")
		return
	}

	// utils.RespondJSON
	utils.RespondJSON(w, http.StatusOK, todo)
}

func GetByStatus(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "error in getting user id")
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		utils.RespondError(w, http.StatusBadRequest, err, "status is required")
		return
	}

	if status != "active" && status != "completed" && status != "incomplete" {
		utils.RespondError(w, http.StatusBadRequest, err, "invalid status")
		return
	}

	todos, err := database.Gettodobystatus(userID, status)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve todos")
		return
	}

	utils.RespondJSON(w, http.StatusOK, todos)
}
func GetBysession(w http.ResponseWriter, r *http.Request) {

	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to get user id")
		return
	}

	todos, err := database.GetTodosByUser(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve todos")
		return
	}

	utils.RespondJSON(w, http.StatusOK, todos)
}

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	sessionToken, err := utils.GetSessionToken(r)
	if err != nil {
		http.Error(w, "unauthorized"+err.Error(), http.StatusUnauthorized)
		return
	}

	var req models.CreateTodoRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "error in parsing body")
		return
	}

	var deadline *time.Time
	if req.Deadline != "" {
		t, err := time.Parse(time.RFC3339, req.Deadline)
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, err, "error in parsing deadline")
			return
		}
		deadline = &t
	}

	todoId, err := database.CreateTodoBySession(
		sessionToken,
		req.Title,
		req.Description,
		deadline,
	)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Fail tp create session")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, map[string]string{
		"todoId": todoId,
	})
}
func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	sessionToken, err := utils.GetSessionToken(r)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "unauthorized")
		return
	}

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	err = database.DeleteTodoBySession(sessionToken, id)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Fail tp delete session")
		return
	}

	utils.RespondJSON(w, http.StatusNoContent, nil)
}
func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	sessionToken, err := utils.GetSessionToken(r)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "unauthorized")
		return
	}

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	err = utils.ParseBody(r.Body, &req)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "error in parsing body")
		return
	}

	err = database.UpdateTodoBySession(

		sessionToken,
		id,
		req.Title,
		req.Description,
		req.Status,
	)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Fail tp update session")
	}

	utils.RespondJSON(w, http.StatusNoContent, nil)
}
func UpcomingTodos(w http.ResponseWriter, r *http.Request) {
	sessionToken, err := utils.GetSessionToken(r)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "unauthorized")
		return
	}

	var days *int
	if d := r.URL.Query().Get("days"); d != "" {
		v, err := strconv.Atoi(d)
		if err != nil || v <= 0 {
			utils.RespondError(w, http.StatusBadRequest, err, "invalid days")
			return
		}
		days = &v
	}

	todos, err := database.GetUpcomingTodosBySession(
		sessionToken,
		days,
	)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve todos")
		return
	}

	json.NewEncoder(w).Encode(todos)
}

func getSessionToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", errors.New("authorization header missing")
	}

	if !strings.HasPrefix(auth, "Bearer ") {
		return "", errors.New("invalid authorization format")
	}

	return strings.TrimPrefix(auth, "Bearer "), nil
}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, err, "Unauthorized")
		return
	}

	err = database.ArchiveUser(userID) // Fixed spelling
	if err != nil {

		utils.RespondError(w, http.StatusNotFound, err, "User not found")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	sessionToken, err := utils.GetSessionToken(r)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "unauthorized")
		return
	}

	userID, err := database.GetUserIDBySession(sessionToken)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve user id")
		return
	}

	err = database.DeleteSession(sessionToken)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to delete session")
		return
	}

	err = database.UpdateUserTimestamp(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err,
			"Failed to update user timestamp")
		return
	}

	utils.RespondJSON(w, http.StatusNoContent, nil)
}
func GetTodo(w http.ResponseWriter, r *http.Request) {

	userID, err := utils.GetUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	page := 1
	limit := 10

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			page = v
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	offset := (page - 1) * limit

	todos, err := database.GetTodosByUserId(userID, limit, offset)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to get user")
		return
	}

	utils.RespondJSON(w, http.StatusOK, todos)
}
func SearchTodo(w http.ResponseWriter, r *http.Request) {
	sessionToken, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "unauthorized")
		return
	}
	title := r.URL.Query().Get("title")
	if title == "" {
		utils.RespondError(w, http.StatusBadRequest, err, "title is required")
		return
	}
	todos, err := database.Search(sessionToken, title)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to search todos")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

//func GetTodos(w http.ResponseWriter, r *http.Request) {
//
//	userID := r.URL.Query().Get("token_id")
//
//	if userID == "" {
//		http.Error(w, "user_id is required", http.StatusBadRequest)
//		return
//	}
//
//	todos, err := database.GetTodosByUser(database2.Todo, userID)
//	if err != nil {
//		http.Error(w, "failed to fetch todos", http.StatusInternalServerError)
//		return
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	json.NewEncoder(w).Encode(todos)
//}

// func FilterTodos(w http.ResponseWriter, r *http.Request) {
//
//		userID := r.URL.Query().Get("user_id")
//		daysStr := r.URL.Query().Get("days")
//
//		if userID == "" || daysStr == "" {
//			http.Error(w, "user_id and days are required", http.StatusBadRequest)
//			return
//		}
//
//		days, err := strconv.Atoi(daysStr)
//		if err != nil || days <= 0 {
//			http.Error(w, "invalid days value", http.StatusBadRequest)
//			return
//		}
//
//		todos, err := database.GetUpcomingTodos(database2.Todo, userID, days)
//		if err != nil {
//			http.Error(w, "failed to fetch todos", http.StatusInternalServerError)
//			return
//		}
//
//		w.Header().Set("Content-Type", "application/json")
//		json.NewEncoder(w).Encode(todos)
//	}
//
//	func DeleteTodo(w http.ResponseWriter, r *http.Request) {
//		idStr := chi.URLParam(r, "id")
//		if idStr == "" {
//			http.Error(w, "todo id is required", http.StatusBadRequest)
//			return
//		}
//
//		todoID, err := strconv.Atoi(idStr)
//		if err != nil {
//			http.Error(w, "invalid todo id", http.StatusBadRequest)
//			return
//		}
//
//		err = database.DeleteTodo(database2.Todo, todoID)
//		if err != nil {
//			http.Error(w, "failed to delete todo", http.StatusInternalServerError)
//			return
//		}
//
//		w.WriteHeader(http.StatusOK)
//		w.Write([]byte("todo deleted"))
//	}
func GetTodosHandler(w http.ResponseWriter, r *http.Request) {

	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionToken := strings.TrimPrefix(auth, "Bearer ")

	userID, err := database.GetUserIDBySession(sessionToken)
	if err != nil {
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return
	}

	todos, err := database.GetTodosByUser(userID)
	if err != nil {
		http.Error(w, "could not fetch todos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func UpdateTodoStatus(w http.ResponseWriter, r *http.Request) {

	todoIDStr := r.URL.Query().Get("todo_id")

	if todoIDStr == "" {
		http.Error(w, "todo_id is required", http.StatusBadRequest)
		return
	}

	todoID, err := strconv.Atoi(todoIDStr)
	if err != nil {
		http.Error(w, "invalid todo_id", http.StatusBadRequest)
		return
	}

	var req models.UpdateTodoStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": errs,
		})

	}

	err = database.UpdateTodoStatus(todoID, req.Title, req.Description, req.Deadline)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"Message":     "Done",
		"Title":       req.Title,
		"Description": req.Description,
		"Deadline":    req.Deadline,
	})
}

func lLogout(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Authorization")

	if sessionToken == "" {
		http.Error(w, "session token required", http.StatusUnauthorized)
		return
	}

	err := database.DeleteSession(sessionToken)
	if err != nil {
		http.Error(w, "failed to logout", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("logged out successfully"))
}

func GetTodos(w http.ResponseWriter, r *http.Request) {
	sessionToken, err := utils.GetSessionToken(r)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "unauthorized")
		return
	}

	todos, err := database.GetTodosBySession(database2.Todo, sessionToken)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve todos")
		return
	}

	json.NewEncoder(w).Encode(todos)
}
