package main

import (
	"fmt"
)

type StudentClass map[string]*string

func main() {

	classA := "Math"
	classPointerA := &classA

	classB := "Science"
	classPointerB := &classB

	students := StudentClass{
		"John":     classPointerA,
		"Jane":     classPointerB,
		"Kianoosh": classPointerB,
	}

	for key, value := range students {
		fmt.Printf("Student: %s, Class: %s\n", key, *value)
	}

	classJohn := students["Kianoosh"]
	fmt.Println(classJohn)
}
