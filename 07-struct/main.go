package main

import "fmt"

type User struct {
	ID     int
	Name   string
	Family string
	Score  int
}

func (u *User) IncreaseScore(value int) {
	u.Score += value
}

func main() {
	// todo add github repo 
	var v int = 10
	var p *int

	fmt.Println(v)
	fmt.Println(&v)
	p = &v
	fmt.Println(p)
	fmt.Println(*p)

	// user1 := User{
	// 	ID:     1,
	// 	Name:   "John",
	// 	Family: "Doe",
	// 	Score:  18,
	// }

	// user1.IncreaseScore(5)

	// fmt.Println(user1)
}
