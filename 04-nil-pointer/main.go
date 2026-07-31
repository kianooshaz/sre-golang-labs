package main

import "fmt"

type Config struct {
	Database string
}

func main() {

	var config *Config

	fmt.Println(config.Database)
}

// func main(){

// 	var config *Config

// 	if config == nil {

// 		fmt.Println("Config is missing")

// 		return
// 	}

// 	fmt.Println(config.Database)

// }
