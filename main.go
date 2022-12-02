package main

import (
	"os"

	"github.com/alex-berlin-tv/radio-ingest/config"
	"github.com/alex-berlin-tv/radio-ingest/daemon"
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
				Name:   "record",
				Usage:  "saves notification bodies to a JSON for further testing",
				Action: recordCmd,
				Flags: []cli.Flag{
					&cli.PathFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "path to config file",
					},
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
			{
				Name:   "test-run",
				Usage:  "test run command with an existing notification",
				Action: testRunCmd,
				Flags: []cli.Flag{
					&cli.PathFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "path to config file",
					},
					&cli.PathFlag{
						Name:    "input",
						Aliases: []string{"i"},
						Usage:   "existing notification body from a JSON file",
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
	cfg := config.ConfigFromDefaults()
	return cfg.ToJSON(ctx.Path("output"))
}

func recordCmd(ctx *cli.Context) error {
	cfg, err := config.ConfigFromJSON(ctx.Path("config"))
	if err != nil {
		return err
	}
	dmn := daemon.NewDaemon(*cfg)
	dmn.Record(ctx.Path("output"))
	return nil
}

func runCmd(ctx *cli.Context) error {
	cfg, err := config.ConfigFromJSON(ctx.Path("config"))
	if err != nil {
		return err
	}
	dmn := daemon.NewDaemon(*cfg)
	dmn.Run()
	return nil
}

func testRunCmd(ctx *cli.Context) error {
	cfg, err := config.ConfigFromJSON(ctx.Path("config"))
	if err != nil {
		return err
	}
	dmn := daemon.NewDaemon(*cfg)
	return dmn.TestRun(ctx.Path("input"))
}
