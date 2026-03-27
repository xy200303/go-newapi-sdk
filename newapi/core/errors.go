package core

import (
	"errors"
	"fmt"
	"strings"
)

var ErrNotFound = errors.New("newapi: resource not found")

type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}

	switch {
	case e.StatusCode > 0 && strings.TrimSpace(e.Message) != "":
		return fmt.Sprintf("newapi API error (%d): %s", e.StatusCode, e.Message)
	case e.StatusCode > 0:
		return fmt.Sprintf("newapi API error (%d)", e.StatusCode)
	case strings.TrimSpace(e.Message) != "":
		return "newapi API error: " + e.Message
	default:
		return "newapi API error"
	}
}
