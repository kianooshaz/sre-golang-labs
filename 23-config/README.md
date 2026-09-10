# 23 — Config from YAML with koanf

Koanf is a config loader where the file format and storage are plugins: here we read `config.yaml` with the yaml provider, and also layer environment variables on top.

```bash
go run .
APP_SERVER__PORT=9090 APP_LOG__LEVEL=debug go run .
```

| Concept            | Where                                                            |
| ------------------ | ---------------------------------------------------------------- |
| Load a YAML file   | `k.Load(file.Provider("config.yaml"), yaml.Parser())`            |
| Dotted-key access  | `k.String("server.host")`, `k.Int`, `k.Bool`                     |
| Env var overrides  | `env.Provider("APP_", ...)`, `APP_SERVER__PORT` -> `server.port` |
| Typed struct       | `k.Unmarshal("", &cfg)` with `koanf` struct tags                 |
| Durations & lists  | `read_timeout: 5s` -> `time.Duration`, `workers:` -> slice       |

Docs: https://github.com/knadh/koanf
