package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/wibecoderr/ToDo/Model"
	"github.com/wibecoderr/ToDo/database"
)

func IsUserRegistered(email string) (bool, error) {
	query := `
       SELECT count(*) > 0
       FROM users
       WHERE email = TRIM(LOWER($1))
       and archived_at IS NULL  
    `

	var exists bool
	err := database.Todo.Get(&exists, query, email)
	return exists, err
}
func CreateUser(email, name, password, phoneNumber string, age int) (string, error) {
	query := `
       INSERT INTO users (email, name, password, phone_number, age)
       VALUES (TRIM(LOWER($1)), $2, $3, $4, $5)
       RETURNING uid
    `

	var userID string
	err := database.Todo.Get(&userID, query, email, name, password, phoneNumber, age)
	return userID, err
}

func GetUserByEmail(email string) (string, string, error) {
	query := `
		SELECT uid, password
		FROM users
		WHERE email = TRIM(LOWER($1))
		AND archived_at IS NULL
	`

	var result struct {
		UID      string `db:"uid"`
		Password string `db:"password"`
	}
	err := database.Todo.Get(&result, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil
		}
		return "", "", err
	}
	return result.UID, result.Password, nil
}
func GetFilteredTodos(sessionToken string, status *string, days int) ([]models.Todo, error) {
	userID, err := GetUserIDBySession(sessionToken)
	if err != nil {
		return nil, err
	}

	var endTime *time.Time
	if days != 0 {
		t := time.Now().AddDate(0, 0, days)
		endTime = &t
	}

	query := `
    SELECT id,
           title,
           description,
           status
    FROM todo
    WHERE user_id = $1
      AND archived_at IS NULL
      AND ($2::text IS NULL OR status = $2::status_enum)
      AND ($3::timestamptz IS NULL OR (
         deadline IS NOT NULL
         AND deadline BETWEEN NOW() AND $3
      ))
    ORDER BY deadline desc
    `

	var todos []models.Todo
	err = database.Todo.Select(&todos, query, userID, status, endTime)
	if err != nil {
		return nil, err
	}

	if todos == nil {
		todos = []models.Todo{}
	}

	return todos, nil
}
func GetArchivedUserByEmail(email string) (string, error) {
	query := `
		SELECT uid
		FROM users
		WHERE email = TRIM(LOWER($1))
		AND archived_at IS NOT NULL
	`
	var userID string
	err := database.Todo.Get(&userID, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return userID, nil
}

// func CreateTodo(db sqlx.Ext, userID, title, description string, deadline *time.Time) error {
//
//		query := `
//			INSERT INTO todo (user_id, title, description, deadline)
//			VALUES ($1, $2, $3, $4)
//			RETURNING id
//		`
//
//		err := db.QueryRowx(query, userID, title, description, deadline).Scan(&todoID)
//		if err != nil {
//			return 0, err
//		}
//
//		return nil
//	}

func RestoreArchivedUser(userID string) error {
	query := `
		UPDATE users
		SET archived_at = NULL,
		    updated_at = now()
		WHERE uid = $1
	`
	_, err := database.Todo.Exec(query, userID)
	return err
}
func GetTodosByUser(userID string) ([]models.Todo, error) {
	query := `
SELECT t.*
FROM todo t
JOIN user_session us ON us.user_id = t.user_id
WHERE us.session_token = $1
  AND us.expires_at > now()
ORDER BY t.created_at DESC;`

	var todos []models.Todo
	err := database.Todo.Select(&todos, query, userID)
	return todos, err
}

func GetUpcomingTodos(
	db sqlx.Ext,
	userID string,
	days int,
) ([]models.Todo, error) {

	query := `
		SELECT *
		FROM todo
		WHERE user_id = $1
		  AND deadline IS NOT NULL
		  AND deadline >= NOW()
		  AND deadline < NOW() + ($2 || ' days')::INTERVAL
		ORDER BY status, created_at DESC
	`

	var todos []models.Todo
	err := sqlx.Select(db, &todos, query, userID, days)
	if err != nil {
		return nil, err
	}

	return todos, nil
}

func UpdateTodoStatus(todoID int, title, description, status string) error {
	sql := `UPDATE todo SET title = $1, description = $2, status = $3 WHERE id = $4`
	_, err := database.Todo.Exec(sql, title, description, status, todoID)
	return err
}

func DeleteTodo(db sqlx.Ext, todoID int) error {
	query := `
        UPDATE todo 
        SET archived_at = now(),
            updated_at = now()
        WHERE id = $1 
        AND archived_at IS NULL
    `
	result, err := db.Exec(query, todoID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("todo not found or already deleted")
	}

	return nil
}

func CreateSession(db sqlx.Ext, userID string) (string, error) {
	token := uuid.New().String()

	query := `
		INSERT INTO user_session (user_id, session_token, expires_at)
		VALUES ($1, $2, now() + interval '24 hours')
	`

	_, err := db.Exec(query, userID, token)
	if err != nil {
		return "", err
	}

	return token, nil
}
func DeleteSession(token string) error {
	sql := "delete from user_session where session_token = $1"
	_, err := database.Todo.Exec(sql, token)
	if err != nil {
		return err
	}
	return nil

}
func GetTodoByID(todoID, userID string) (models.Todo, error) {
	sql := `SELECT * FROM todo 
            WHERE id = $1 
            AND user_id = $2 
            AND archived_at IS NULL`

	var todo models.Todo
	err := database.Todo.Get(&todo, sql, todoID, userID)
	if err != nil {
		return models.Todo{}, errors.New("todo not found")

	}
	return todo, nil
}
func Gettodobystatus(userID, status string) ([]models.Todo, error) {
	sql := `SELECT t.*
		FROM todo t
	JOIN user_session us ON us.user_id = t.user_id
	WHERE us.session_token = $1
	AND t.status = $2
	ORDER BY t.created_at DESC;`

	var todo []models.Todo
	err := database.Todo.Select(&todo, sql, userID, status)
	if err != nil {
		return nil, err
	}
	return todo, nil
}
func GetUserBySession(db sqlx.Ext, sessionToken string) (string, error) {
	query := `
		SELECT user_id
		FROM user_session
		WHERE session_token = $1
		  AND expires_at > now()
	`

	var userID string
	err := sqlx.Get(db, &userID, query, sessionToken)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func GetUserIDBySession(token string) (string, error) {
	var userID string

	query := `
		SELECT user_id
		FROM user_session
		WHERE session_token = $1
		  AND expires_at > now()
	`

	err := database.Todo.Get(&userID, query, token)
	return userID, err
}
func DeleteTodoBySession(token string, id int) error {
	query := `
        UPDATE todo
        SET archived_at = now(),
            updated_at = now()
        WHERE id = $2
          AND archived_at IS NULL
          AND user_id = (
            SELECT user_id 
            FROM user_session 
            WHERE session_token = $1
            AND expires_at > now()
          )
    `
	result, err := database.Todo.Exec(query, token, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("todo not found or unauthorized")
	}

	return nil
}

func UpdateTodoBySession(token string, id int, title, desc, status string) error {
	query := `
	UPDATE todo t
	SET 
		title = $3,
		description = $4,
		status = $5,
		updated_at = now()
	FROM user_session us
	WHERE t.id = $2
	  AND t.user_id = us.user_id
	  AND us.session_token = $1
	  AND us.expires_at > now()
	`

	result, err := database.Todo.Exec(query, token, id, title, desc, status)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("todo not found or unauthorized")
	}

	return nil
}
func GetUpcomingTodosBySession(
	token string,
	days *int,
) ([]models.Todo, error) {

	query := `
	SELECT t.*
	FROM todo t
	JOIN user_session us ON us.user_id = t.user_id
	WHERE us.session_token = $1
	  AND (
		$2 IS NULL
		OR t.deadline < now() + ($2 || ' days')::interval
	  )
	ORDER BY t.created_at DESC
	`

	var todos []models.Todo
	err := database.Todo.Select(&todos, query, token, days)
	return todos, err
}
func GetTodosBySession(db sqlx.Ext, token string) ([]models.Todo, error) {
	var todos []models.Todo
	query := `
	SELECT t.*
	FROM todo t
	JOIN user_session us ON us.user_id = t.user_id
	WHERE us.session_token = $1
	ORDER BY t.created_at DESC
	`
	err := sqlx.Select(db, &todos, query, token)
	return todos, err
}
func CreateTodoBySession(
	token, title, description string,
	deadline *time.Time,
) (string, error) {

	query := `
    INSERT INTO todo (user_id, title, description, deadline)
    SELECT user_id, $2, $3, $4
    FROM user_session
    WHERE session_token = $1
    RETURNING id
    `

	var todoID string
	err := database.Todo.Get(&todoID, query, token, title, description, deadline)
	return todoID, err
}
func ArchiveUser(userID string) error {
	query := `
        UPDATE users
        SET archived_at = now(),
            updated_at = now()
        WHERE uid = $1 
        AND archived_at IS NULL
    `
	result, err := database.Todo.Exec(query, userID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("user not found or already deleted")
	}

	return nil
}
func UpdateUserTimestamp(userID string) error {
	query := `UPDATE users SET updated_at = now() WHERE uid = $1`
	_, err := database.Todo.Exec(query, userID)
	return err
}
func GetTodosByUserId(userId string, limit, offset int) ([]models.Todo, error) {

	// safety (optional but good)
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT title, description, status,
		       deadline, created_at
		FROM todo
		WHERE user_id = $1
		  AND archived_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var todos []models.Todo

	err := database.Todo.Select(&todos, query, userId, limit, offset)
	if err != nil {
		return nil, err
	}

	// return empty slice instead of nil
	if todos == nil {
		todos = []models.Todo{}
	}

	return todos, nil
}

func SearchTodos(userID, title string) ([]models.Todo, error) {
	sql := `SELECT title, description, created_at, deadline 
            FROM todo 
            WHERE user_id = $1
              AND archived_at IS NULL
              AND title ILIKE '%' || $2 || '%'
            ORDER BY created_at DESC`

	var todos []models.Todo
	err := database.Todo.Select(&todos, sql, userID, title)
	if err != nil {
		return nil, err
	}

	if todos == nil {
		todos = []models.Todo{}
	}

	return todos, nil
}
