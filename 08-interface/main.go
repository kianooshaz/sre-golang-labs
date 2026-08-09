package main

import "fmt"

type Server interface {
	Up()
	Down()
	Status() string
}

type linuxServer struct {
	CPU string
	RAM string
}

func (l *linuxServer) Up() {
	fmt.Println("Linux server is up")
}

func (l *linuxServer) Down() {
	fmt.Println("Linux server is down")
}

func (l *linuxServer) Status() string {
	return "Linux server status: OK"
}

type windowsServer struct {
	CPU string
	RAM string
}

func (w *windowsServer) Up() {
	fmt.Println("Windows server is up")
}

func (w *windowsServer) Down() {
	fmt.Println("Windows server is down")
}

func (w *windowsServer) Status() string {
	return "Windows server status: OK"
}

func main() {
	var myServer Server

	myServer = &linuxServer{
		CPU: "Intel",
		RAM: "16GB",
	}

	// ...

	myServer = &windowsServer{
		CPU: "Intel",
		RAM: "16GB",
	}

	fmt.Println(myServer)
}
