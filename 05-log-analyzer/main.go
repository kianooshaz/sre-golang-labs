package main

import "fmt"

func analyzeLogs(logs []string) map[string]int {

	result := make(map[string]int)

	for _, log := range logs {

		switch log {

		case "ERROR":
			result["error"]++

		case "WARN":
			result["warning"]++

		case "INFO":
			result["info"]++

		default:
			result["unknown"]++

		}

	}

	return result

}

func main() {

	logs := []string{
		"INFO",
		"INFO",
		"ERROR",
		"WARN",
		"ERROR",
	}

	report := analyzeLogs(logs)

	fmt.Println(report)

}
