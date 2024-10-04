package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/AndreZiviani/boundary-fuzzy/internal/keyring"
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/authmethods"
)

type OidcLogin struct {
	boundaryClient *api.Client
	authClient     *authmethods.Client
}

func (o *OidcLogin) Execute(ctx context.Context, methodId string) error {
	result, err := o.authClient.Authenticate(ctx, methodId, "start", nil)
	if err != nil {
		if apiErr := api.AsServerError(err); apiErr != nil {
			return apiErr
		}
		return err
	}

	startResp := new(authmethods.OidcAuthMethodAuthenticateStartResponse)
	if err := json.Unmarshal(result.GetRawAttributes(), startResp); err != nil {
		return err
	}

	fmt.Printf("Open the following URL in your browser to authenticate:\n%s\n", startResp.AuthUrl)

	var watchCode int
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(1500 * time.Millisecond):
				result, err = o.authClient.Authenticate(ctx, methodId, "token", map[string]any{
					"token_id": startResp.TokenId,
				})
				if err != nil {
					if apiErr := api.AsServerError(err); apiErr != nil {
						fmt.Println(apiErr)
						return
					}
					fmt.Println(err)
					return
				}
				if result.GetResponse().StatusCode() == http.StatusAccepted {
					// Nothing yet -- circle around.
					continue
				}

				return
			}
		}
	}()
	wg.Wait()

	if watchCode != 0 {
		return fmt.Errorf("Error watching for code: %d", watchCode)
	}
	if result == nil {
		return fmt.Errorf("No response from the server")
	}

	return keyring.SaveTokenToKeyring(result)
}
