package client

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/AndreZiviani/boundary-fuzzy/internal/keyring"
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/authtokens"
	"github.com/hashicorp/boundary/api/sessions"
)

func NewBoundaryClient(ctx context.Context) (*api.Client, *authtokens.AuthToken, error) {
	boundaryAddr := os.Getenv("BOUNDARY_ADDR")
	if boundaryAddr == "" {
		return nil, nil, fmt.Errorf("environment variable BOUNDARY_ADDR is not set")
	}

	config, _ := api.DefaultConfig()
	config.Addr = boundaryAddr
	boundaryClient, err := api.NewClient(config)
	if err != nil {
		return nil, nil, err
	}

	boundaryToken, err := keyring.GetBoundaryToken()
	if err != nil {
		// could not retrieve token from keyring
		return boundaryClient, nil, err
	}

	if time.Now().Before(boundaryToken.ExpirationTime) {
		boundaryClient.SetToken(boundaryToken.Token)
		sessionsClient := sessions.NewClient(boundaryClient)
		_, err = sessionsClient.List(ctx, "global", sessions.WithRecursive(true))
		if err != nil {
			// token is invalid
			boundaryClient.SetToken("")
			boundaryToken = nil
		}
	} else {
		boundaryToken = nil
	}

	return boundaryClient, boundaryToken, nil
}
