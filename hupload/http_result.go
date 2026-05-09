package main

import (
	"encoding/json"
	"net/http"
)

type APIResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func setJSONHeaders(w http.ResponseWriter) {
	headers := w.Header()
	headers.Set("Content-Type", "application/json; charset=utf-8")
	headers.Set("Cache-Control", "no-store, no-cache, must-revalidate")
	headers.Set("Pragma", "no-cache")
	headers.Set("Expires", "0")
	headers.Set("Vary", "Cookie, Authorization")
}

func writeError(w http.ResponseWriter, code int, msg string) {
	setJSONHeaders(w)
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(APIResult{Status: "error", Message: msg})
}

func writeSuccessJSON(w http.ResponseWriter, body any) {
	setJSONHeaders(w)
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func writeSuccess(w http.ResponseWriter, message string) {
	writeSuccessJSON(w, APIResult{Status: "success", Message: message})
}
