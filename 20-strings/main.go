// Learning strings in Go: convert, compare, build, search, split, trim.
// Run: go run .
package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func main() {
	// 1. Build strings (strings.Builder)
	build()

	// 2. Convert: string <-> number, string <-> []byte
	convert()

	// 3. Compare
	compare()

	// 4. Search and check
	search()

	// 5. Split and join
	splitJoin()

	// 6. Trim and change case
	trimCase()

	// 7. Replace and repeat
	replaceRepeat()

	// 8. Runes: length vs byte count
	runes()
}

func build() {
	// a) simple concatenation — fine for a few strings
	s := "server-" + "01" + " up"
	fmt.Println(s) // server-01 up

	// b) fmt.Sprintf — when you need formatting
	s = fmt.Sprintf("%s:%d", "localhost", 8080)
	fmt.Println(s) // localhost:8080

	// c) strings.Join — build from a slice
	parts := []string{"web-01", "web-02", "db-01"}
	fmt.Println(strings.Join(parts, ",")) // web-01,web-02,db-01

	// d) strings.Builder — best in loops, no reallocations per step
	var b strings.Builder
	b.Grow(32) // optional: pre-allocate if you know the rough size
	for i := 1; i <= 3; i++ {
		fmt.Fprintf(&b, "server-%d ", i)
	}
	fmt.Println(b.String()) // server-1 server-2 server-3

	// e) strings.Builder with WriteString/WriteByte
	b.Reset()
	b.WriteString("GET ")
	b.WriteByte('/')
	b.WriteString("healthz")
	fmt.Println(b.String()) // GET /healthz

	// f) bytes.Buffer — same idea, also works as an io.Writer for libs
	var buf bytes.Buffer
	buf.WriteString("log: ")
	buf.WriteString("disk full")
	fmt.Println(buf.String()) // log: disk full

	// g) []byte with append — lowest level, rare in normal code
	data := []byte("line1")
	data = append(data, '\n')
	data = append(data, "line2"...)
	fmt.Println(string(data)) // line1\nline2
}

func convert() {
	n := 42
	s := fmt.Sprintf("%d", n)  // int -> string
	i, _ := strconv.Atoi("25") // string -> int
	f, _ := strconv.ParseFloat("3.14", 64)
	s = strconv.Itoa(25)

	bs := []byte("hello") // string -> []byte
	str := string(bs)     // []byte -> string

	fmt.Println(s, i, f, str)
}

func compare() {
	fmt.Println("abc" == "abc")                      // true
	fmt.Println(strings.Compare("a", "b"))           // -1 (a before b)
	fmt.Println(strings.EqualFold("Hello", "HELLO")) // true, ignore case
}

func search() {
	s := "user-service is running on port 8080"
	fmt.Println(strings.Contains(s, "running")) // true
	fmt.Println(strings.HasPrefix(s, "user"))   // true
	fmt.Println(strings.HasSuffix(s, "8080"))   // true
	fmt.Println(strings.Index(s, "port"))       // position, -1 if missing
	fmt.Println(strings.Count(s, "e"))          // how many times
}

func splitJoin() {
	csv := "web-01,web-02,db-01"
	parts := strings.Split(csv, ",")
	fmt.Println(parts)                      // [web-01 web-02 db-01]
	fmt.Println(strings.Join(parts, " | ")) // web-01 | web-02 | db-01

	// Split by whitespace, any amount
	fmt.Println(strings.Fields("cpu 82.5% mem 61%")) // [cpu 82.5% mem 61%]
}

func trimCase() {
	logLine := "  WARN: disk almost full \n"
	fmt.Println(strings.TrimSpace(logLine)) // "WARN: disk almost full"
	fmt.Println(strings.ToUpper("events"))  // EVENTS
	fmt.Println(strings.ToLower("WARN"))    // warn
}

func replaceRepeat() {
	s := "2026-09-02"
	fmt.Println(strings.ReplaceAll(s, "-", "/")) // 2026/09/02
	fmt.Println(strings.Replace(s, "-", "/", 1)) // 2026/09-02
	fmt.Println(strings.Repeat("=", 10))         // ==========
}

func runes() {
	s := "héllo"
	fmt.Println(len(s))         // 6: bytes, not letters
	fmt.Println(len([]rune(s))) // 5: real letters
	for i, r := range s {       // range gives runes
		fmt.Println(i, string(r))
	}
}
