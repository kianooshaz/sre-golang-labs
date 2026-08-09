package main

import (
	"http-server/service"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", service.HealthHandler)
	mux.HandleFunc("GET /services", service.GetServicesHandler)
	mux.HandleFunc("GET /services/{id}", service.GetServiceHandler)
	mux.HandleFunc("POST /services", service.CreateServiceHandler)
	mux.HandleFunc("DELETE /services/{id}", service.DeleteServiceHandler)

	handler := service.LoggingMiddleware(mux)

	log.Println("Server listening on :8080")

	log.Fatal(http.ListenAndServe(":8080", handler))
}
