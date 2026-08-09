package service

import (
	"time"
)

var Services = map[int]Service{
	1: {
		ID:        1,
		Name:      "Auth Service",
		URL:       "https://auth.example.com",
		Status:    "UP",
		CreatedAt: time.Now(),
	},
	2: {
		ID:        2,
		Name:      "Payment Service",
		URL:       "https://payment.example.com",
		Status:    "DOWN",
		CreatedAt: time.Now(),
	},
}

var nextID = 3
