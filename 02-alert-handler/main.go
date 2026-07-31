package main

import "fmt"

func handleAlert(alert string) {

	switch alert {

	case "CPU_HIGH":
		fmt.Println("Scale up application")

	case "DISK_FULL":
		fmt.Println("Clean old logs")

	case "SERVICE_DOWN":
		fmt.Println("Restart service")

	default:
		fmt.Println("Unknown alert")
	}
}

func main() {

	// map
	alerts := map[string]int{
		"CPU_HIGH":     5,
		"DISK_FULL":    2,
		"SERVICE_DOWN": 1,
	}

	// loop over map
	for alert, count := range alerts {

		fmt.Println(alert, count)

		handleAlert(alert)
	}

}
