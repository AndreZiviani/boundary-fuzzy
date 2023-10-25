package cli

import (
	"fmt"
	"os"

	"github.com/AndreZiviani/boundary-fuzzy/internal/target"

	//"github.com/AndreZiviani/boundary-fuzzy/internal/ssh"
	//"github.com/AndreZiviani/boundary-fuzzy/internal/chart"
	//"github.com/AndreZiviani/boundary-fuzzy/internal/config"
	//"github.com/AndreZiviani/boundary-fuzzy/internal/ssm"
	//"github.com/AndreZiviani/boundary-fuzzy/internal/sso"
	"github.com/urfave/cli/v2"
)

var (
	version string
)

func Run() error {

	if len(version) == 0 {
		version = "Unknown version, manually compiled from git?"
	}

	flags := []cli.Flag{
		&cli.BoolFlag{Name: "verbose", Usage: "Log debug messages"},
	}

	app := &cli.App{
		Flags:       flags,
		Name:        "boundary-fuzzy",
		Usage:       "https://github.com/AndreZiviani/boundary-fuzzy",
		UsageText:   "boundary-fuzzy [global options] command [command options] [arguments...]",
		Version:     version,
		HideVersion: false,
		Commands: []*cli.Command{
			target.Command(),
			/*
				ssh.Command(),
				config.Command(),
				chart.Command(),
				sso.Command(),
				ssm.Command(),
			*/
		},
		EnableBashCompletion: true,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)

		os.Exit(1)
	}

	return err
}
