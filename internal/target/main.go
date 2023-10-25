package target

import (
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/targets"
	"github.com/urfave/cli/v2"
)

type List struct {
	ScopeName      string
	ScopeID        string
	boundaryClient *api.Client
	targetClient   *targets.Client
}

type Connect struct {
	boundaryClient *api.Client
	targetClient   *targets.Client
}

func Command() *cli.Command {
	command := cli.Command{
		Name:  "target",
		Usage: "Target Utilities",
		Subcommands: []*cli.Command{
			{
				Name:  "connect",
				Usage: "Connect to a target",
				Flags: []cli.Flag{},
				Action: func(c *cli.Context) error {
					connect := NewConnect()

					return connect.Execute(c.Context)
				},
			},
		},
	}

	return &command
}
