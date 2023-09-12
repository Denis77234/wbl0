package handler

import (
	"net/http"
)

func checkMethods(r *http.Request, allowedMethods []string) bool {
	for _, method := range allowedMethods {
		if r.Method == method {
			return true
		}
	}

	return false
}
