package main

import (
	"fmt"
	"time"

	"github.com/mshafiee/jalali"
)

func main() {
	now := time.Now()

	fmt.Println("Current Year:", now.Year())
	fmt.Println("Current Month:", now.Month())
	fmt.Println("Current Day:", now.Day())

	fmt.Println("Next Day: ", now.Add(24*time.Hour))
	fmt.Println("Next Min: ", now.Add(time.Minute))
	fmt.Println("Next Year: ", now.AddDate(1, 0, 0))
	fmt.Println("Next Day: ", now.AddDate(0, 0, 1))

	expiredTime := time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC)
	fmt.Println("Specific Time: ", expiredTime)

	fmt.Println("Is Expired: ", expiredTime.Before(time.Now()))

	futureTime := time.Date(2027, 10, 1, 12, 0, 0, 0, time.Local)
	print(time.Until(futureTime))
	fmt.Println()
	fmt.Println("Until Future time:", time.Until(futureTime))
	fmt.Println("human readable format", futureTime.Format(time.RFC3339))

	jalaliTime := jalali.ToJalali(futureTime)
	if jalaliTime.Day() != 1 {
		return
	}

	// TODO
}
