# 20 — Strings in Go

One file, eight basics: build, convert, compare, search, split/join, trim/case, replace/repeat, runes.

```bash
go run .
```

| Section       | Functions                                                                                   |
| ------------- | ------------------------------------------------------------------------------------------- |
| 1. Build      | `+`, `fmt.Sprintf`, `strings.Join`, `strings.Builder`, `bytes.Buffer`, `append` on `[]byte` |
| 2. Convert    | `strconv.Atoi`, `strconv.ParseFloat`, `fmt.Sprintf`, `[]byte(...)`                          |
| 3. Compare    | `==`, `strings.Compare`, `strings.EqualFold`                                                |
| 4. Search     | `strings.Contains`, `HasPrefix`, `HasSuffix`, `Index`, `Count`                              |
| 5. Split/Join | `strings.Split`, `strings.Join`, `strings.Fields`                                           |
| 6. Trim/Case  | `strings.TrimSpace`, `ToUpper`, `ToLower`                                                   |
| 7. Replace    | `strings.ReplaceAll`, `strings.Repeat`                                                      |
| 8. Runes      | `len` (bytes) vs `[]rune` (letters), `range` over runes                                     |
