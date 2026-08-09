package main

import (
	"encoding/json"
	"os"
)

type RequestCreatePod struct {
	Name2     string `json:"name"`
	CPU       string `json:"cpu,omitempty"`
	memory    string `json:"memory"`
	Namespace string `json:"namespace"`
}

type ResponseCreatePod struct {
	Message string `json:"message"`
}

func main() {
	req := RequestCreatePod{
		Name2:     "my-pod",
		memory:    "256Mi",
		Namespace: "default",
	}

	json.NewEncoder(os.Stdout).Encode(req)
}
