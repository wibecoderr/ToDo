package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/wibecoderr/ToDo"
	utils "github.com/wibecoderr/ToDo"
	database2 "github.com/wibecoderr/ToDo/database"
	database "github.com/wibecoderr/ToDo/database/dbhelper"
	models "github.com/wibecoderr/ToDo/model"
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
		utils.RespondError(w, http.StatusConflict, nil, "User already exists")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to hash password")
		return
	}
	var jwtToken string
	//  err = database2.Tx(func(tx *sqlx.Tx) error {
	err = database2.Tx(func(tx *sqlx.Tx) error {
		userID, err := database.CreateUser(
			tx,
			req.Email,
			req.Name,
			hashedPassword,
			req.PhoneNumber,
			req.Age)

		if err != nil {
			return err
		}

		// Create session for auto-login

		sessionID, err := database.CreateSession(tx, userID)
		if err != nil {
			return err
		}
		// jwt token
		jwtToken, err = utils.GenerateJWT(userID, sessionID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create user")
		return
	}

	// utils - Respond JSON
	utils.RespondJSON(w, http.StatusCreated, map[string]string{
		"token":   jwtToken,
		"message": "registered and logged in",
	})
}
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var (
		req      models.LoginRequest
		jwtToken string
	)
	// parse body -- validate -- check user exists -- decoding -- transcation --session -- jwt -- encode

	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse body")
		return
	}
	if err := utils.ValidateStruct(req); err != nil {
		utils.RespondValidationError(w, err)
		return
	}
	// checking user is in database or not
	userId, password, err := database.GetUserByEmail(req.Email)
	if userId == "" {
		utils.RespondError(w, http.StatusNotFound, nil, "User not found")
		return
	}
	if userId == "" {
		utils.RespondError(w, http.StatusNotFound, nil, "User not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to get user")
		return
	}
	// check password

	if !utils.CheckPasswordHash(req.Password, password) {
		utils.RespondError(w, http.StatusUnauthorized, nil, "User passwords do not match")
		return
	}
	// password checked now
	// transcastion

	err = database2.Tx(func(tx *sqlx.Tx) error {
		sessionID, err := database.CreateSession(tx, userId)
		if err != nil {
			return err
		}
		jwtToken, err = utils.GenerateJWT(userId, sessionID)
		if err != nil {
			return err
		}
		return nil

	})
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, nil, "Failed to create session")
		return
	}
	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"token":   jwtToken,
		"message": "logged in",
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
		utils.RespondError(w, http.StatusInternalServerError, err, "Fail to create session")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, map[string]int{
		"todoId": todoId,
	})
}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	userCTX := utils.UserContext(r)
	userID := userCTX.UserId
	todoID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "invalid  parameter")
		return
	}
	err = database.DeleteTodoById(userID, todoID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Fail to  delete session")
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
		return
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
		utils.RespondError(w, http.StatusInternalServerError, err, "Fail to update session")
		return
	}

	utils.RespondJSON(w, http.StatusNoContent, nil)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	userCTX := utils.UserContext(r)
	userID := userCTX.UserId
	err := database2.Tx(func(tx *sqlx.Tx) error {
		return database.ArchiveUser(tx, userID)
	})
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "user not found or failed to delete")
		return
	}
	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {

	userCtx := utils.UserContext(r)
	sessionToken := userCtx.SessionId
	userID := userCtx.UserId

	err := database2.Tx(func(tx *sqlx.Tx) error {

		if err := database.DeleteSession(tx, sessionToken); err != nil {
			return err
		}

		if err := database.UpdateUserTimestamp(tx, userID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		utils.RespondError(
			w,
			http.StatusInternalServerError,
			err,
			"Logout failed",
		)
		return
	}

	utils.RespondJSON(w, http.StatusNoContent, nil)
}
