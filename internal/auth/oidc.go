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
	"github.com/hashicorp/boundary/api/authtokens"
	"github.com/pkg/browser"
)

func (a *Auth) OIDCLogin(ctx context.Context, methodId string) (*authtokens.AuthToken, error) {
	result, err := a.authClient.Authenticate(ctx, methodId, "start", nil)
	if err != nil {
		if apiErr := api.AsServerError(err); apiErr != nil {
			return nil, apiErr
		}
		return nil, err
	}

	startResp := new(authmethods.OidcAuthMethodAuthenticateStartResponse)
	if err := json.Unmarshal(result.GetRawAttributes(), startResp); err != nil {
		return nil, err
	}

	err = browser.OpenURL(startResp.AuthUrl)
	if err != nil {
		fmt.Printf("Failed to automatically open authentication link, please open this link:\n\n%s\n", startResp.AuthUrl)
	} else {
		fmt.Printf("Please finish the authentication process on your browser\n")
	}

	var watchCode int
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(1500 * time.Millisecond):
				result, err = a.authClient.Authenticate(ctx, methodId, "token", map[string]any{
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
		return nil, fmt.Errorf("Error watching for code: %d", watchCode)
	}
	if result == nil {
		return nil, fmt.Errorf("No response from the server")
	}

	_ = keyring.SaveTokenToKeyring(result)
	return result.GetAuthToken()
}
