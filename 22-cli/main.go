// Learning CLI tools with urfave/cli: commands, flags, args, and help text.
// Run: go run . greet --name Sre   |   go run . http get https://example.com
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "sretool",
		Usage: "a tiny Swiss-army knife for SRE chores",
		// Version is shown by `sretool --version`
		Version: "1.0.0",
		// Global flags: available to every subcommand
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"V"}, // note: lowercase -v is taken by --version
				Usage:   "log more detail",
			},
		},
		Commands: []*cli.Command{
			greetCmd(),
			checkCmd(),
			httpCmd(),
		},
	}

	// cli parses os.Args, dispatches to the right command, prints help on error.
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// greetCmd: `sretool greet --name Kianoosh --shout` — flags with a default.
func greetCmd() *cli.Command {
	return &cli.Command{
		Name:  "greet",
		Usage: "say hello to someone",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Value: "world", // default value
				Usage: "who to greet",
			},
			&cli.IntFlag{
				Name:  "times",
				Value: 1,
				Usage: "how many times",
			},
			&cli.BoolFlag{
				Name:  "shout",
				Usage: "print in upper case",
			},
		},
		Action: func(c *cli.Context) error {
			name := c.String("name")
			if c.Bool("shout") {
				name = fmt.Sprint(name, "!") + " (LOUD)"
			}
			for i := 0; i < c.Int("times"); i++ {
				fmt.Printf("hello, %s\n", name)
			}
			// Show a global flag: Bool on the parent context
			if c.Bool("verbose") {
				fmt.Printf("[verbose] greeted %q %d times\n", name, c.Int("times"))
			}
			return nil
		},
	}
}

// checkCmd: a positional argument — `sretool check web-01`.
func checkCmd() *cli.Command {
	return &cli.Command{
		Name:      "check",
		Usage:     "simulate a health check of a host",
		ArgsUsage: "<host>",
		Action: func(c *cli.Context) error {
			host := c.Args().First() // first positional arg
			if host == "" {
				return cli.ShowSubcommandHelp(c) // print help when arg is missing
			}
			fmt.Printf("checking %s ... OK (200 in 12ms)\n", host)
			return nil
		},
	}
}

// httpCmd: nested subcommands — `sretool http get <url>`.
func httpCmd() *cli.Command {
	return &cli.Command{
		Name:  "http",
		Usage: "tiny http client",
		Subcommands: []*cli.Command{
			{
				Name:      "get",
				Usage:     "GET a url",
				ArgsUsage: "<url>",
				Action: func(c *cli.Context) error {
					url := c.Args().First()
					if url == "" {
						return fmt.Errorf("usage: sretool http get <url>")
					}
					start := time.Now()
					client := &http.Client{Timeout: 5 * time.Second}
					resp, err := client.Get(url) //nolint:noctx — tiny example
					if err != nil {
						return err
					}
					defer resp.Body.Close()
					fmt.Printf("%s %s -> %d %s (%s)\n",
						c.Command.Name, url, resp.StatusCode, resp.Status, time.Since(start).Round(time.Millisecond))
					return nil
				},
			},
		},
	}
}
