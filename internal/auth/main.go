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
		Action: func(c *cli.Context) error {
			auth := &Auth{}

			return auth.Execute(c.Context)
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

	return "", fmt.Errorf("Primary auth method not found in global scope")
}

func (a *Auth) Execute(ctx context.Context) error {
	boundaryClient, _, err := client.NewBoundaryClient()
	a.boundaryClient = boundaryClient

	a.authClient = authmethods.NewClient(a.boundaryClient)
	pri, err := a.getPrimaryAuthMethodId(ctx)
	if err != nil {
		return err
	}

	switch {
	case strings.HasPrefix(pri, globals.OidcAuthMethodPrefix):
		oidc := &OidcLogin{
			boundaryClient: a.boundaryClient,
			authClient:     a.authClient,
		}
		return oidc.Execute(ctx, pri)
	case strings.HasPrefix(pri, globals.PasswordAuthMethodPrefix):
		// todo
		return fmt.Errorf("Password login is not implemented")
	case strings.HasPrefix(pri, globals.LdapAuthMethodPrefix):
		// todo
		return fmt.Errorf("LDAP login is not implemented")
	}

	return fmt.Errorf("Unknown auth method type")
}
