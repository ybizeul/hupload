package middleware

import (
	"encoding/json"
	"net/http"
)

type APIResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(APIResult{Status: "error", Message: msg})
}
