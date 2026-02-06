package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	models "github.com/wibecoderr/ToDo/Model"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func ParseBody(body io.Reader, out interface{}) error {
	err := json.NewDecoder(body).Decode(out)
	if err != nil {
		return err
	}

	return nil
}

// EncodeJSONBody writes the JSON body to response writer
func EncodeJSONBody(resp http.ResponseWriter, data interface{}) error {
	return json.NewEncoder(resp).Encode(data)
}

// RespondJSON sends the interface as a JSON
func RespondJSON(w http.ResponseWriter, statusCode int, body interface{}) {
	w.WriteHeader(statusCode)
	if body != nil {
		if err := EncodeJSONBody(w, body); err != nil {
			logrus.Errorf("Failed to respond JSON with error: %+v", err)
		}
	}
}

var validate = validator.New()

func ValidateStruct(data interface{}) map[string]string {
	err := validate.Struct(data)
	if err == nil {
		return nil
	}
	errors := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		errors[e.Field()] = e.Tag()
	}

	return errors
}

func RespondError(w http.ResponseWriter, statusCode int, err error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	resp := models.Error{
		Error:      errStr,
		StatusCode: statusCode,
		Message:    message,
	}

	if encodeErr := json.NewEncoder(w).Encode(resp); encodeErr != nil {
		fmt.Printf("failed to encode error response: %v\n", encodeErr)
	}
}
func RespondValidationError(w http.ResponseWriter, validationErrors map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := map[string]interface{}{
		"errors":      validationErrors,
		"status_code": http.StatusBadRequest,
		"message":     "Validation failed",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("failed to encode validation error response: %v\n", err)
	}
}
