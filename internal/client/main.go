package client

import (
	"fmt"
	"os"

	"github.com/AndreZiviani/boundary-fuzzy/internal/keyring"
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/authtokens"
)

func NewBoundaryClient() (*api.Client, *authtokens.AuthToken, error) {
	boundaryAddr := os.Getenv("BOUNDARY_ADDR")
	if boundaryAddr == "" {
		return nil, nil, fmt.Errorf("Environment variable BOUNDARY_ADDR is not set")
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
	boundaryClient.SetToken(boundaryToken.Token)

	return boundaryClient, boundaryToken, nil
}
