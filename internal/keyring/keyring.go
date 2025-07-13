package keyring

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/hashicorp/boundary/api/authmethods"
	"github.com/hashicorp/boundary/api/authtokens"
	nkeyring "github.com/jefferai/keyring"
	zkeyring "github.com/zalando/go-keyring"
)

const (
	NoneKeyring          = "none"
	AutoKeyring          = "auto"
	WincredKeyring       = "wincred"
	PassKeyring          = "pass"
	KeychainKeyring      = "keychain"
	SecretServiceKeyring = "secret-service"

	DefaultTokenName = "default"
	LoginCollection  = "login"
	PassPrefix       = "HashiCorp_Boundary"

	EnvToken        = "BOUNDARY_TOKEN"
	EnvTokenName    = "BOUNDARY_TOKEN_NAME"
	EnvKeyringType  = "BOUNDARY_KEYRING_TYPE"
	StoredTokenName = "HashiCorp Boundary Auth Token"
)

func GetBoundaryToken() (*authtokens.AuthToken, error) {
	token := os.Getenv(EnvToken)
	if len(token) == 0 {
		keyringType, tokenName, err := discoverKeyringTokenInfo()
		if err != nil {
			return nil, err
		}
		authToken, err := readTokenFromKeyring(keyringType, tokenName)
		if err != nil {
			return nil, err
		}
		return authToken, nil
	} else {
		return &authtokens.AuthToken{
			Token: token,
		}, nil
	}
}

func discoverKeyringTokenInfo() (string, string, error) {
	// Stops the underlying library from invoking a dbus call that ends up
	// freezing some systems
	os.Setenv("DISABLE_KWALLET", "1")

	tokenName := DefaultTokenName

	// Set so we can look it up later when printing out curl strings
	os.Setenv(EnvTokenName, tokenName)

	var foundKeyringType bool
	keyringType := os.Getenv(EnvKeyringType)
	switch runtime.GOOS {
	case "windows":
		switch keyringType {
		case AutoKeyring, WincredKeyring, PassKeyring:
			foundKeyringType = true
			if keyringType == AutoKeyring {
				keyringType = WincredKeyring
			}
		}
	case "darwin":
		switch keyringType {
		case AutoKeyring, KeychainKeyring, PassKeyring:
			foundKeyringType = true
			if keyringType == AutoKeyring {
				keyringType = KeychainKeyring
			}
		}
	default:
		switch keyringType {
		case AutoKeyring, SecretServiceKeyring, PassKeyring:
			foundKeyringType = true
			if keyringType == AutoKeyring {
				keyringType = PassKeyring
			}
		}
	}

	if !foundKeyringType {
		return "", "", fmt.Errorf("given keyring type %q is not valid, or not valid for this platform", EnvKeyringType)
	}

	var available bool
	switch keyringType {
	case WincredKeyring, KeychainKeyring:
		available = true
	case PassKeyring, SecretServiceKeyring:
		avail := nkeyring.AvailableBackends()
		for _, a := range avail {
			if keyringType == string(a) {
				available = true
			}
		}
	}

	if !available {
		return "", "", fmt.Errorf("keyring type %q is not available on this machine. For help with setting up keyrings, see https://www.boundaryproject.io/docs/api-clients/cli", keyringType)
	}

	os.Setenv(EnvKeyringType, keyringType)

	return keyringType, tokenName, nil
}

func readTokenFromKeyring(keyringType, tokenName string) (*authtokens.AuthToken, error) {
	var token string
	var err error

	switch keyringType {
	case NoneKeyring:
		return nil, nil

	case WincredKeyring, KeychainKeyring:
		token, err = zkeyring.Get(StoredTokenName, tokenName)
		if err != nil {
			if err == zkeyring.ErrNotFound {
				return nil, errors.Join(fmt.Errorf("no saved credential found"), err)
			} else {
				return nil, errors.Join(fmt.Errorf("error reading auth token from keyring"), err)
			}
		}

	default:
		krConfig := nkeyring.Config{
			LibSecretCollectionName: LoginCollection,
			PassPrefix:              PassPrefix,
			AllowedBackends:         []nkeyring.BackendType{nkeyring.BackendType(keyringType)},
		}

		var kr nkeyring.Keyring
		kr, err = nkeyring.Open(krConfig)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("error opening keyring"), err)
		}

		var item nkeyring.Item
		item, err = kr.Get(tokenName)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("error reading auth token from keyring"), err)
		}

		token = string(item.Data)
	}

	if token != "" {
		tokenBytes, err := base64.RawStdEncoding.DecodeString(token)
		switch {
		case err != nil:
			return nil, errors.Join(fmt.Errorf("error base64-unmarshaling stored token from system credential store"), err)
		case len(tokenBytes) == 0:
			return nil, errors.New("zero length token after decoding stored token from system credential store")
		default:
			var authToken authtokens.AuthToken
			if err := json.Unmarshal(tokenBytes, &authToken); err != nil {
				return nil, errors.Join(fmt.Errorf("error unmarshaling stored token information after reading from system credential store"), err)
			} else {
				return &authToken, nil
			}
		}
	}
	return nil, err
}

func TokenIdFromToken(token string) (string, error) {
	split := strings.Split(token, "_")
	if len(split) < 3 {
		return "", errors.New("unexpected stored token format")
	}
	return strings.Join(split[0:2], "_"), nil
}

func SaveTokenToKeyring(result *authmethods.AuthenticateResult) error {
	token := new(authtokens.AuthToken)
	if err := json.Unmarshal(result.GetRawAttributes(), token); err != nil {
		return err
	}

	keyringType, tokenName, err := discoverKeyringTokenInfo()
	if err != nil {
		return err
	} else if keyringType != "none" && tokenName != "none" && keyringType != "" && tokenName != "" {
		marshaled, err := json.Marshal(token)
		if err != nil {
			return err
		}
		switch keyringType {
		case "wincred", "keychain":
			if err := zkeyring.Set(StoredTokenName, tokenName, base64.RawStdEncoding.EncodeToString(marshaled)); err != nil {
				return err
			}
		default:
			krConfig := nkeyring.Config{
				LibSecretCollectionName: LoginCollection,
				PassPrefix:              PassPrefix,
				AllowedBackends:         []nkeyring.BackendType{nkeyring.BackendType(keyringType)},
			}

			kr, err := nkeyring.Open(krConfig)
			if err != nil {
				return err
			}

			if err := kr.Set(nkeyring.Item{
				Key:  tokenName,
				Data: []byte(base64.RawStdEncoding.EncodeToString(marshaled)),
			}); err != nil {
				return err
			}
		}

		fmt.Printf("token %q saved to keyring %q\n", tokenName, keyringType)

	}

	// fmt.Printf("Token: %s\n", token.Token)
	return nil
}
