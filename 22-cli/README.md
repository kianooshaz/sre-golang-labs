# 22 — CLI tools with urfave/cli

One file, four basics: app + global flags, subcommands with flags, positional args, nested subcommands.

```bash
go run . --help
go run . greet --name Sre --shout
go run . -V greet --name Kianoosh --times 2
go run . check web-01
go run . http get https://example.com
```

| Concept            | Where                                                            |
| ------------------ | ---------------------------------------------------------------- |
| App & global flags | `cli.App{Flags: ...}`, `-v` / `--verbose`                        |
| Subcommand flags   | `StringFlag` default value, `IntFlag`, `BoolFlag`, `c.String(..)`|
| Positional args    | `c.Args().First()`, `ArgsUsage`, help on missing arg             |
| Nested commands    | `Subcommands` (`http get <url>` with a real HTTP request)        |
| Help & version     | `--help` auto-generated, `--version` from `App.Version`          |

Docs: https://github.com/urfave/cli
