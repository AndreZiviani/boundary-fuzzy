package client

import (
	"fmt"
	"os"

	"github.com/hashicorp/boundary/api"
)

func NewBoundaryClient() (*api.Client, error) {
	boundaryAddr := os.Getenv("BOUNDARY_ADDR")
	if boundaryAddr == "" {
		return nil, fmt.Errorf("Environment variable BOUNDARY_ADDR is not set")
	}

	config, _ := api.DefaultConfig()
	config.Addr = boundaryAddr
	boundaryClient, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	boundaryToken, err := getBoundaryToken()
	if err != nil {
		return nil, err
	}
	boundaryClient.SetToken(boundaryToken.Token)

	return boundaryClient, nil
}
