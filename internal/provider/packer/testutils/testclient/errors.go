package testclient

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

type ErrorWithCode interface {
	error
	Code() int
}

func isAlreadyExistsError(e ErrorWithCode) bool {
	switch e.Code() {
	case int(codes.AlreadyExists), http.StatusConflict:
		return true
	default:
		return false
	}
}
