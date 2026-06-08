package core

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// HTTPError represents a structured error response.
type HTTPError struct {
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

// BindAndValidate populates the object with data from the request and validates it.
func BindAndValidate(r *http.Request, obj any) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("obj must be a non-nil pointer")
	}

	el := v.Elem()
	if el.Kind() != reflect.Struct {
		return errors.New("obj must point to a struct")
	}

	// 1. Bind JSON Body if present
	if r.ContentLength > 0 && strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
			return &HTTPError{Message: "Invalid JSON body"}
		}
	}

	// 2. Bind Query & Path via Reflection
	t := el.Type()
	q := r.URL.Query()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := el.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		var val string

		// Query tag
		if queryTag := field.Tag.Get("query"); queryTag != "" {
			if qVal := q.Get(queryTag); qVal != "" {
				val = qVal
			}
		}

		// Path tag (overrides query if both are somehow present and mapped)
		if pathTag := field.Tag.Get("path"); pathTag != "" {
			if pVal := r.PathValue(pathTag); pVal != "" {
				val = pVal
			}
		}

		if val != "" {
			switch field.Type.Kind() {
			case reflect.String:
				fieldValue.SetString(val)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
					fieldValue.SetInt(intVal)
				}
			case reflect.Bool:
				if boolVal, err := strconv.ParseBool(val); err == nil {
					fieldValue.SetBool(boolVal)
				}
			}
		}
	}

	// 3. Validate Struct
	if err := validate.Struct(obj); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			errMap := make(map[string]string)
			for _, fieldError := range validationErrors {
				errMap[fieldError.Field()] = fieldError.Tag()
			}
			return &HTTPError{Message: "Validation failed", Errors: errMap}
		}
		return err
	}

	return nil
}
