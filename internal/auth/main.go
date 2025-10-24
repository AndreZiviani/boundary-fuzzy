package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/AndreZiviani/boundary-fuzzy/internal/client"
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/authmethods"
	"github.com/hashicorp/boundary/globals"
	"github.com/urfave/cli/v2"
)

type Auth struct {
	boundaryClient *api.Client
	authClient     *authmethods.Client
}

func Command() *cli.Command {
	command := cli.Command{
		Name:  "auth",
		Usage: "Auth Utilities",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Usage:   "force reauthentication",
				Value:   false,
				Aliases: []string{"f"},
			},
		},
		Action: func(c *cli.Context) error {
			auth := &Auth{}

			return auth.Execute(c)
		},
	}

	return &command
}

func (a *Auth) getPrimaryAuthMethodId(ctx context.Context) (string, error) {
	authMethods, err := a.authClient.List(ctx, "global")
	if err != nil {
		return "", err
	}

	for _, authMethod := range authMethods.GetItems() {
		if authMethod.IsPrimary {
			return authMethod.Id, nil
		}
	}

	return "", fmt.Errorf("primary auth method not found in global scope")
}

func (a *Auth) Execute(c *cli.Context) error {
	_, err := Login(c.Context, c.Bool("force"))
	return err
}

func Login(ctx context.Context, force bool) (string, error) {
	boundaryClient, token, err := client.NewBoundaryClient(ctx)
	auth := &Auth{
		boundaryClient: boundaryClient,
		authClient:     authmethods.NewClient(boundaryClient),
	}

	if err != nil {
		return "", err
	}

	if !force {
		if token != nil {
			fmt.Printf("Using cached credentials\n")
			return token.Token, nil
		}
	}

	pri, err := auth.getPrimaryAuthMethodId(ctx)
	if err != nil {
		return "", err
	}

	switch {
	case strings.HasPrefix(pri, globals.OidcAuthMethodPrefix):
		result, err := auth.OIDCLogin(ctx, pri)
		if err != nil || result == nil {
			return "", err
		}
		return result.Token, nil

	case strings.HasPrefix(pri, globals.PasswordAuthMethodPrefix):
		// todo
		return "", fmt.Errorf("password login is not implemented")
	case strings.HasPrefix(pri, globals.LdapAuthMethodPrefix):
		// todo
		return "", fmt.Errorf("LDAP login is not implemented")
	}

	return "", fmt.Errorf("unknown auth method type")

}
