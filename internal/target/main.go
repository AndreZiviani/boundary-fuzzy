package target

import (
	"github.com/AndreZiviani/boundary-fuzzy/internal/client"
	"github.com/AndreZiviani/boundary-fuzzy/internal/target/tui"
	"github.com/hashicorp/boundary/api/targets"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	command := cli.Command{
		Name:  "target",
		Usage: "Target Utilities",
		Subcommands: []*cli.Command{
			{
				Name:   "connect",
				Usage:  "Connect to a target",
				Flags:  []cli.Flag{},
				Action: TargetTui,
			},
		},
	}

	return &command
}

func TargetTui(c *cli.Context) error {
	boundaryClient, token, err := client.NewBoundaryClient(c.Context)
	if err != nil {
		return err
	}

	targetClient := targets.NewClient(boundaryClient)

	targets, err := targetClient.List(c.Context, "global", targets.WithRecursive(true))
	if err != nil {
		return err
	}

	tui.Tui(c.Context, targets, boundaryClient, token)
	return nil
}
