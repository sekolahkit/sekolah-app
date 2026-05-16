package validator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

func DecodeJSON(r *http.Request, dst interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var numericRegex = regexp.MustCompile(`^[0-9]+$`)

func Required(field, value string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Message: "wajib diisi"}
	}
	return nil
}

func MinLength(field, value string, min int) *ValidationError {
	if len(value) > 0 && len(value) < min {
		return &ValidationError{Field: field, Message: fmt.Sprintf("minimal %d karakter", min)}
	}
	return nil
}

func MaxLength(field, value string, max int) *ValidationError {
	if len(value) > max {
		return &ValidationError{Field: field, Message: fmt.Sprintf("maksimal %d karakter", max)}
	}
	return nil
}

func Email(field, value string) *ValidationError {
	if value != "" && !emailRegex.MatchString(value) {
		return &ValidationError{Field: field, Message: "format email tidak valid"}
	}
	return nil
}

func Numeric(field, value string) *ValidationError {
	if value != "" && !numericRegex.MatchString(value) {
		return &ValidationError{Field: field, Message: "harus berupa angka"}
	}
	return nil
}

func InList(field, value string, allowed []string) *ValidationError {
	if value == "" {
		return nil
	}
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return &ValidationError{Field: field, Message: fmt.Sprintf("harus salah satu dari: %s", strings.Join(allowed, ", "))}
}

func Collect(errs ...*ValidationError) ValidationErrors {
	var result ValidationErrors
	for _, err := range errs {
		if err != nil {
			result = append(result, *err)
		}
	}
	return result
}
