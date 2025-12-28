package errx

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Err        error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code
}

func (e *AppError) Unwrap() error { return e.Err }

func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status}
}

func Wrap(err error, code, message string, status int) *AppError {
	if err == nil {
		return New(code, message, status)
	}
	return &AppError{Code: code, Message: message, HTTPStatus: status, Err: err}
}

func As(err error) (*AppError, bool) {
	var app *AppError
	if errors.As(err, &app) && app != nil {
		if app.HTTPStatus == 0 {
			app.HTTPStatus = http.StatusInternalServerError
		}
		if app.Code == "" {
			app.Code = "INTERNAL_ERROR"
		}
		if app.Message == "" {
			app.Message = "internal error"
		}
		return app, true
	}
	return nil, false
}

