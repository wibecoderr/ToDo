package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/wibecoderr/ToDo"
	utils "github.com/wibecoderr/ToDo"
	database2 "github.com/wibecoderr/ToDo/database"
	database "github.com/wibecoderr/ToDo/database/dbhelper"
	"github.com/wibecoderr/ToDo/model"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterUserRequest

	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse body")
		return
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}

	// Check if user already exists
	exists, err := database.IsUserRegistered(req.Email)
	if err != nil {
		// utils
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to check if user exists")
		return
	}
	if exists {
		utils.RespondError(w, http.StatusBadRequest, err, "User already exists")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to hash password")
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
	//// for delete
	//if userID == "" {
	//	// check wheether user exist or not is exists tell user to contact
	//	archivedUserID, err := database.GetArchivedUserByEmail(req.Email)
	//	if err != nil {
	//		utils.RespondError(w, http.StatusInternalServerError, err, "server error")
	//		return
	//	}
	//
	//	if archivedUserID != "" {
	//		utils.RespondError(w, http.StatusBadRequest, err, "user_id is not registered")
	//		return
	//	}
	//
	//	utils.RespondError(w, http.StatusBadRequest, err, "invalid user")
	//}

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
	userCTX := utils.UserContext(r)
	userID := userCTX.UserId

	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		if s != "active" && s != "completed" && s != "incomplete" {
			utils.RespondError(w, http.StatusBadRequest, nil, "invalid status")
			return
		}
		status = &s
	}

	days := 0
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err != nil || d < 0 {
			utils.RespondError(w, http.StatusBadRequest, err, "invalid days parameter")
			return
		}
		days = d
	}

	page := 1
	limit := 10

	if p := r.URL.Query().Get("page"); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil || v < 1 {
			utils.RespondError(w, http.StatusBadRequest, err, "invalid page")
			return
		}
		page = v
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		v, err := strconv.Atoi(l)
		if err != nil || v < 1 || v > 100 {
			utils.RespondError(w, http.StatusBadRequest, err, "invalid limit")
			return
		}
		limit = v
	}

	offset := (page - 1) * limit

	todos, err := database.GetFilteredTodos(
		userID,
		status,
		days,
		limit,
		offset,
	)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to retrieve todos")
		return
	}

	utils.RespondJSON(w, http.StatusOK, todos)
}

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	userCTX := utils.UserContext(r)
	userID := userCTX.UserId

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
		userID,
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
	userCTX := utils.UserContext(r)
	userID := userCTX.UserId
	todoID, err := strconv.Atoi(chi.URLParam(r, "")) //error check
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "invalid  parameter")
		return
	}
	err = database.DeleteTodoById(userID, todoID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Fail tp delete session")
		return
	}

	utils.RespondJSON(w, http.StatusNoContent, nil)
}

func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	// no need of session todoID
	userCtx := utils.UserContext(r)
	userID := userCtx.UserId

	todoID, err := strconv.Atoi(chi.URLParam(r, "todoID"))
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "invalid  parameter")
	}

	var reqBody models.TodoRequest

	err = utils.ParseBody(r.Body, &reqBody)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "error in parsing body")
		return
	}
	if errs := utils.ValidateStruct(reqBody); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}
	err = database.UpdateTodoById(
		userID,
		todoID,
		reqBody.Title,
		reqBody.Description,
		reqBody.Status,
	)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Fail tp update session")
		return
	}

	utils.RespondJSON(w, http.StatusNoContent, nil)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	userCTX := utils.UserContext(r)
	userID := userCTX.UserId

	err := database.ArchiveUser(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "User not found") //500
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {

	userCtx := utils.UserContext(r)
	SessionToken := userCtx.SessionId
	userID := userCtx.UserId
	//userID, err := database.GetUserIDBySession(sessionToken)
	//if err != nil {
	//	utils.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve user id")
	//	return
	//}

	err := database.DeleteSession(SessionToken) // transaction
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
