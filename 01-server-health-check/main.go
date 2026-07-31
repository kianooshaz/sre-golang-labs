package main

import "fmt"

func checkServerStatus(cpuUsage int) string {
	if cpuUsage > 90 {
		return "CRITICAL"
	}

	return "OK"
}

func main() {
	// fetch the CPU usages of servers from a monitoring system
	cpuUsagesOfServers := map[string]int{
		"server1": 75,
		"server2": 85,
		"server3": 95,
	}

	for serverName, cpuUsage := range cpuUsagesOfServers {
		status := checkServerStatus(cpuUsage)
		fmt.Printf("Server: %s, Status: %s\n", serverName, status)
	}
}
