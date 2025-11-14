package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registra rutas HTTP mínimas (usadas por quienes consuman el paquete
// directamente vía net/http). En la aplicación principal usamos Gin, por eso aquí
// dejamos solo /health para compatibilidad.
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/health", health).Methods(http.MethodGet)
	// NOTA: los endpoints de autenticación y usuarios se registran en `cmd/api/main.go`
	// usando Gin. Mantener estas rutas aquí causaba referencias a símbolos que no
	// existen (Login, GetProfile) en forma de handlers net/http.
}
