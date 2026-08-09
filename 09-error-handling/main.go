package main

import "fmt"

func Div(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero is not allowed")
	}

	return a / b, nil
}

func main() {
	result, err := Div(10, 0)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Result:", result)
}
