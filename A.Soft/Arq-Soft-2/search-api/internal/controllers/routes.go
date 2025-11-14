package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/health", health).Methods(http.MethodGet)
	// TODO: add handlers
}

func health(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}
