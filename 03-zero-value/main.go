package main

import "fmt"

type Server struct {
	Name   string
	IP     string
	CPU    int
	Status bool
}

func main() {

	var srv Server

	fmt.Println(srv)

	srv.Name = "database-01"
	srv.IP = "10.0.0.5"

	fmt.Println(srv)

}
