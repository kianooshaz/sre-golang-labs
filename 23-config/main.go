// Learning config loading with koanf: read a YAML file, access values, and
// override with environment variables.
// Run: go run .
package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config is the typed shape we unmarshal into — apps prefer structs over
// dotted-key lookups scattered around the code. Note the tags: koanf's
// Unmarshal matches keys via `koanf` tags.
type Config struct {
	Server struct {
		Host string        `koanf:"host"`
		Port int           `koanf:"port"`
		TLS  bool          `koanf:"tls"`
		Read time.Duration `koanf:"read_timeout"`
	} `koanf:"server"`

	Log struct {
		Level  string `koanf:"level"`
		Format string `koanf:"format"`
	} `koanf:"log"`

	Workers []struct {
		Name    string `koanf:"name"`
		Queue   int    `koanf:"queue_size"`
		Enabled bool   `koanf:"enabled"`
	} `koanf:"workers"`
}

func main() {
	k := koanf.New(".")

	// 1. Load config.yaml — the parser is a plugin, so the same code works
	//    for JSON/TOML by swapping the Parser.
	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		log.Fatalf("load config.yaml: %v", err)
	}

	// 2. Dotted-key access — quick lookups without a struct.
	fmt.Println("host :", k.String("server.host"))
	fmt.Println("port :", k.Int("server.port"))
	fmt.Println("tls  :", k.Bool("server.tls"))
	fmt.Println("level:", k.String("log.level"))

	// 3. Environment variable override — APP_SERVER__HOST beats the file.
	//    `__` in the env name maps to a `.` in the key path:
	//    APP_SERVER__HOST -> server.host, APP_LOG__LEVEL -> log.level.
	k.Load(env.Provider("APP_", ".", func(s string) string {
		return envKey(strings.TrimPrefix(s, "APP_"))
	}), nil)

	// 4. Unmarshal the whole tree into a typed struct.
	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		log.Fatalf("unmarshal: %v", err)
	}

	fmt.Printf("\nfinal config:\n%+v\n", cfg)
	fmt.Println("read timeout:", cfg.Server.Read, "-> as ms:", cfg.Server.Read.Milliseconds())
	fmt.Println("worker[0]:", cfg.Workers[0].Name, "queue", cfg.Workers[0].Queue)
}

// envKey turns SERVER__PORT into "server.port": lowercase, `__` -> `.`
func envKey(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && s[i+1] == '_' {
			out = append(out, '.')
			i++
			continue
		}
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out = append(out, c)
	}
	return string(out)
}
