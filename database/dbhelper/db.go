package database

import (
	"database/sql"
	"errors"

	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/wibecoderr/ToDo/database"
	"github.com/wibecoderr/ToDo/model"
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
func CreateUser(tx *sqlx.Tx, email, name, password, phoneNumber string, age int) (string, error) {
	query := `
       INSERT INTO users (email, name, password, phone_number, age)
       VALUES (TRIM(LOWER($1)), $2, $3, $4, $5)
       RETURNING uid
    `

	var userID string
	err := tx.Get(&userID, query, email, name, password, phoneNumber, age)
	return userID, err
}

func GetUserByEmail(email string) (string, string, error) {
	query := `
		SELECT uid, password
		FROM users
		WHERE email = TRIM(LOWER($1))
		AND archived_at IS NULL
	`

	var request models.Result
	err := database.Todo.Get(&request, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil
		}
		return "", "", err
	}
	return request.UID, request.Password, nil
}
func GetFilteredTodos(
	userID string,
	status *string,
	days int,
	limit int,
	offset int,
) ([]models.Todo, error) {

	var endTime *time.Time
	if days != 0 {
		t := time.Now().AddDate(0, 0, days)
		endTime = &t
	}

	query := `
		SELECT id,
			   title,
			   description,
			   status,
			   deadline
		FROM todo
		WHERE user_id = $1
		  AND archived_at IS NULL
		  AND ($2::text IS NULL OR status = $2::status_enum)
		  AND ($3::timestamptz IS NULL OR (
				deadline IS NOT NULL
				AND deadline BETWEEN NOW() AND $3
		  ))
		ORDER BY deadline DESC
		LIMIT $4 OFFSET $5
	`

	todos := make([]models.Todo, 0)
	err := database.Todo.Select(
		&todos,
		query,
		userID,
		status,
		endTime,
		limit,
		offset,
	)

	return todos, err
}

func CreateSession(db sqlx.Ext, userID string) (string, error) {
	sessionID := uuid.New().String()

	query := `
		INSERT INTO user_session (user_id, session_token, expires_at)
		VALUES ($1, $2, now() + interval '24 hours')
				RETURNING id
	`

	var id int
	err := sqlx.Get(db, &id, query, userID, sessionID)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func DeleteSession(token string) error {
	_, err := database.Todo.Exec(
		`DELETE FROM user_session WHERE session_token = $1`,
		token,
	)
	return err

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
func DeleteTodoById(userid string, todoID int) error {
	query := `
        UPDATE todo
        SET archived_at = now()
        WHERE id = $1
          AND user_id = $2
          AND archived_at IS NULL
    `

	_, err := database.Todo.Exec(query, todoID, userid)
	if err != nil {
		return err
	}

	return nil
}

func UpdateTodoById(userId string, todoId int, title, desc, status string) error {
	query := `
        UPDATE todo
        SET title = $3,
            description = $4,
            status = $5,
            updated_at = now()
        WHERE id = $1
          AND user_id = $2
          AND archived_at IS NULL
    `

	result, err := database.Todo.Exec(query, todoId, userId, title, desc, status)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("todo not found or unauthorized")
	}
	return nil
}

func CreateTodoBySession(userID, title, description string, deadline *time.Time) (int, error) {
	query := `
        INSERT INTO todo (user_id, title, description, deadline)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `

	var todoID int
	err := database.Todo.Get(&todoID, query, userID, title, description, deadline)
	return todoID, err
}

//func ArchiveUser(userID string) error {
//	query := `
//        UPDATE users
//        SET archived_at = now(),
//            updated_at = now()
//        WHERE uid = $1
//        AND archived_at IS NULL
//    `
//
//	_, err := database.Todo.Exec(query, userID)
//	return err
//}

func ArchiveUser(tx *sqlx.Tx, userID string) error {

	_, err := tx.Exec(`
		UPDATE todo
		SET archived_at = now()
		WHERE user_id = $1
		AND archived_at IS NULL
	`, userID)
	if err != nil {
		return err
	}

	// for user
	_, err = tx.Exec(`
		UPDATE users
		SET archived_at = now(),
			updated_at = now()
		WHERE uid = $1 
		AND archived_at IS NULL
	`, userID)
	if err != nil {
		return err
	}
	return nil

}
func UpdateUserTimestamp(userID string) error {
	_, err := database.Todo.Exec(
		`UPDATE users SET updated_at = NOW() WHERE uid = $1`,
		userID,
	)
	return err
}
