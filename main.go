package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "radio-ingest",
		Usage: "handles incoming radio uploads",
		Action: func(ctx *cli.Context) error {
			cli.ShowAppHelp(ctx)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:   "config",
				Usage:  "generates a new config file",
				Action: configCmd,
				Flags: []cli.Flag{
					&cli.PathFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "path to output file",
					},
				},
			},
			{
				Name:   "run",
				Usage:  "runs the daemon",
				Action: runCmd,
				Flags: []cli.Flag{
					&cli.PathFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "path to config file",
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func configCmd(ctx *cli.Context) error {
	cfg := ConfigFromDefaults()
	return cfg.ToJSON(ctx.Path("output"))
}

func runCmd(ctx *cli.Context) error {
	return nil
}
