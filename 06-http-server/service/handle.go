package service

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

var Name = "ali"
var age = 30

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func GetServicesHandler(w http.ResponseWriter, r *http.Request) {
	list := make([]Service, 0, len(Services))

	for _, service := range Services {
		list = append(list, service)
	}

	json.NewEncoder(w).Encode(list)
}

func GetServiceHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	service, ok := Services[id]
	if !ok {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(service)
}

func CreateServiceHandler(w http.ResponseWriter, r *http.Request) {
	var service Service

	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	service.ID = nextID
	service.Status = "UNKNOWN"
	service.CreatedAt = time.Now()

	Services[2] = service
	nextID++

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(service)
}

func DeleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if _, ok := Services[id]; !ok {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}

	delete(Services, id)


	w.WriteHeader(http.StatusNoContent)
}
